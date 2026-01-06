package eval_test

import (
	"context"
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/eval"
)

// ExampleNewStreamingTrajectoryScorer demonstrates real-time trajectory evaluation
// during agent execution. The scorer provides progressive feedback as steps are added.
func Example_streamingTrajectoryScorer() {
	// Create a streaming scorer with expected execution sequence
	scorer := eval.NewStreamingTrajectoryScorer(eval.TrajectoryOptions{
		ExpectedSteps: []eval.ExpectedStep{
			{Type: "tool", Name: "nmap", Required: true},
			{Type: "tool", Name: "nuclei", Required: true},
			{Type: "finding", Name: "", Required: true},
		},
		Mode:          eval.TrajectoryOrderedSubset,
		PenalizeExtra: 0.05, // 5% penalty per extra step
	})

	ctx := context.Background()

	// Simulate progressive agent execution
	trajectory := eval.Trajectory{
		StartTime: time.Now(),
		Steps:     []eval.TrajectoryStep{},
	}

	// Step 1: Agent calls nmap (correct)
	trajectory.Steps = append(trajectory.Steps, eval.TrajectoryStep{
		Type:      "tool",
		Name:      "nmap",
		StartTime: time.Now(),
	})

	result, _ := scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("After nmap: score=%.2f action=%s\n", result.Score, result.Action)

	// Step 2: Agent calls hydra (extra step, not in expected sequence)
	trajectory.Steps = append(trajectory.Steps, eval.TrajectoryStep{
		Type:      "tool",
		Name:      "hydra",
		StartTime: time.Now(),
	})

	result, _ = scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("After hydra: score=%.2f action=%s\n", result.Score, result.Action)

	// Step 3: Agent calls nuclei (correct, continues sequence)
	trajectory.Steps = append(trajectory.Steps, eval.TrajectoryStep{
		Type:      "tool",
		Name:      "nuclei",
		StartTime: time.Now(),
	})

	result, _ = scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("After nuclei: score=%.2f action=%s\n", result.Score, result.Action)

	// Step 4: Agent submits finding (completes sequence)
	trajectory.Steps = append(trajectory.Steps, eval.TrajectoryStep{
		Type:      "finding",
		Name:      "",
		StartTime: time.Now(),
	})

	result, _ = scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("After finding: score=%.2f action=%s confidence=%.2f\n",
		result.Score, result.Action, result.Confidence)

	// Output:
	// After nmap: score=0.33 action=continue
	// After hydra: score=0.28 action=continue
	// After nuclei: score=0.62 action=continue
	// After finding: score=0.95 action=continue confidence=1.00
}

// Example_streamingTrajectoryScorer_exactMatch demonstrates strict sequence matching
// where any deviation is immediately flagged.
func Example_streamingTrajectoryScorer_exactMatch() {
	scorer := eval.NewStreamingTrajectoryScorer(eval.TrajectoryOptions{
		ExpectedSteps: []eval.ExpectedStep{
			{Type: "tool", Name: "nmap", Required: true},
			{Type: "tool", Name: "nuclei", Required: true},
		},
		Mode: eval.TrajectoryExactMatch,
	})

	ctx := context.Background()
	trajectory := eval.Trajectory{StartTime: time.Now()}

	// Correct first step
	trajectory.Steps = []eval.TrajectoryStep{
		{Type: "tool", Name: "nmap", StartTime: time.Now()},
	}
	result, _ := scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("Step 1 correct: action=%s\n", result.Action)

	// Wrong second step - immediate deviation detected
	trajectory.Steps = append(trajectory.Steps, eval.TrajectoryStep{
		Type:      "tool",
		Name:      "hydra", // Expected nuclei
		StartTime: time.Now(),
	})
	result, _ = scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("Step 2 wrong: action=%s\n", result.Action)

	// Output:
	// Step 1 correct: action=continue
	// Step 2 wrong: action=reconsider
}

// Example_streamingTrajectoryScorer_subsetMatch demonstrates flexible matching
// where steps can appear in any order.
func Example_streamingTrajectoryScorer_subsetMatch() {
	scorer := eval.NewStreamingTrajectoryScorer(eval.TrajectoryOptions{
		ExpectedSteps: []eval.ExpectedStep{
			{Type: "tool", Name: "nmap", Required: true},
			{Type: "tool", Name: "nuclei", Required: true},
			{Type: "finding", Name: "", Required: true},
		},
		Mode:          eval.TrajectorySubsetMatch,
		PenalizeExtra: 0.0, // Don't penalize extras in this example
	})

	ctx := context.Background()

	// Steps appear out of order but all required steps are present
	trajectory := eval.Trajectory{
		Steps: []eval.TrajectoryStep{
			{Type: "finding", Name: "", StartTime: time.Now()},   // Last expected
			{Type: "tool", Name: "nuclei", StartTime: time.Now()}, // Second expected
			{Type: "tool", Name: "nmap", StartTime: time.Now()},   // First expected
		},
		StartTime: time.Now(),
	}

	result, _ := scorer.ScorePartial(ctx, trajectory)
	fmt.Printf("Out of order: score=%.2f confidence=%.2f\n", result.Score, result.Confidence)

	// Output:
	// Out of order: score=1.00 confidence=1.00
}
