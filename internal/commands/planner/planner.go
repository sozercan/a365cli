package planner

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// PlannerCmd groups all Planner subcommands.
type PlannerCmd struct {
	Plans  PlansCmd  `cmd:"" help:"Planner plans"`
	Tasks  TasksCmd  `cmd:"" help:"Planner tasks"`
	Goals  GoalsCmd  `cmd:"" help:"Planner goals"`
}

func plannerEndpoint() string {
	return config.Endpoint("planner")
}

// --- Plans ---

// PlansCmd groups plan subcommands.
type PlansCmd struct {
	List   PlansListCmd   `cmd:"" help:"List plans"`
	Get    PlansGetCmd    `cmd:"" help:"Get a plan by ID"`
	Create PlansCreateCmd `cmd:"" help:"Create a plan"`
	Update PlansUpdateCmd `cmd:"" help:"Update a plan"`
}

type PlansListCmd struct {
	Max int `help:"Maximum number of results" default:"50"`
}

func (c *PlansListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "QueryPlans", map[string]any{})
	if err != nil {
		return fmt.Errorf("list plans: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "plans")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("plans", output.PlannerPlanColumns, rows)
}

type PlansGetCmd struct {
	ID string `arg:"" help:"Plan ID"`
}

func (c *PlansGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "GetPlan", map[string]any{"planId": c.ID})
	if err != nil {
		return fmt.Errorf("get plan: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

type PlansCreateCmd struct {
	Title string `arg:"" help:"Plan title"`
}

func (c *PlansCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(plannerEndpoint(), "CreatePlan", fmt.Sprintf("create plan %q", c.Title),
			map[string]any{"action": "planner.create-plan", "title": c.Title})
	}
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "CreatePlan", map[string]any{"title": c.Title})
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Plan created", data)
}

type PlansUpdateCmd struct {
	ID    string `arg:"" help:"Plan ID"`
	Title string `help:"New title" optional:""`
}

func (c *PlansUpdateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(plannerEndpoint(), "UpdatePlan", fmt.Sprintf("update plan %s", c.ID),
			map[string]any{"action": "planner.update-plan", "planId": c.ID})
	}
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	args := map[string]any{"planId": c.ID}
	if c.Title != "" {
		args["title"] = c.Title
	}
	resp, err := client.CallTool(ctx.Ctx, "UpdatePlan", args)
	if err != nil {
		return fmt.Errorf("update plan: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Plan updated", data)
}

// --- Tasks ---

type TasksCmd struct {
	List   TasksListCmd   `cmd:"" help:"List tasks in a plan"`
	Get    TasksGetCmd    `cmd:"" help:"Get a task by ID"`
	Create TasksCreateCmd `cmd:"" help:"Create a task"`
	Update TasksUpdateCmd `cmd:"" help:"Update a task"`
}

type TasksListCmd struct {
	PlanID string `arg:"" help:"Plan ID"`
	Max    int    `help:"Maximum number of results" default:"50"`
}

func (c *TasksListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "QueryTasksInPlan", map[string]any{"planId": c.PlanID})
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	rows := output.ToRows(data, "tasks")
	if rows == nil {
		rows = output.ToRows(data, "value")
	}
	if rows == nil {
		return ctx.Output.PrintItem(data)
	}
	if c.Max > 0 && len(rows) > c.Max {
		rows = rows[:c.Max]
	}
	return ctx.Output.PrintList("tasks", output.PlannerTaskColumns, rows)
}

type TasksGetCmd struct {
	ID string `arg:"" help:"Task ID"`
}

func (c *TasksGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "GetTask", map[string]any{"taskId": c.ID})
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

type TasksCreateCmd struct {
	PlanID string `arg:"" help:"Plan ID"`
	Title  string `arg:"" help:"Task title"`
}

func (c *TasksCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(plannerEndpoint(), "CreateTask", fmt.Sprintf("create task %q in plan %s", c.Title, c.PlanID),
			map[string]any{"action": "planner.create-task", "planId": c.PlanID, "title": c.Title})
	}
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "CreateTask", map[string]any{
		"planId": c.PlanID, "title": c.Title,
	})
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Task created", data)
}

type TasksUpdateCmd struct {
	ID    string `arg:"" help:"Task ID"`
	Title string `help:"New title" optional:""`
}

func (c *TasksUpdateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(plannerEndpoint(), "UpdateTask", fmt.Sprintf("update task %s", c.ID),
			map[string]any{"action": "planner.update-task", "taskId": c.ID})
	}
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	args := map[string]any{"taskId": c.ID}
	if c.Title != "" {
		args["title"] = c.Title
	}
	resp, err := client.CallTool(ctx.Ctx, "UpdateTask", args)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Task updated", data)
}

// --- Goals ---

type GoalsCmd struct {
	List   GoalsListCmd   `cmd:"" help:"List goals in a plan"`
	Get    GoalsGetCmd    `cmd:"" help:"Get a goal by ID"`
	Create GoalsCreateCmd `cmd:"" help:"Create a goal"`
	Update GoalsUpdateCmd `cmd:"" help:"Update a goal"`
}

type GoalsListCmd struct {
	PlanID string `arg:"" help:"Plan ID"`
}

func (c *GoalsListCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "QueryGoalsInPlan", map[string]any{"planId": c.PlanID})
	if err != nil {
		return fmt.Errorf("list goals: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

type GoalsGetCmd struct {
	ID string `arg:"" help:"Goal ID"`
}

func (c *GoalsGetCmd) Run(ctx *commands.Context) error {
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "GetGoal", map[string]any{"goalId": c.ID})
	if err != nil {
		return fmt.Errorf("get goal: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

type GoalsCreateCmd struct {
	PlanID string `arg:"" help:"Plan ID"`
	Title  string `arg:"" help:"Goal title"`
}

func (c *GoalsCreateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(plannerEndpoint(), "CreateGoal", fmt.Sprintf("create goal %q in plan %s", c.Title, c.PlanID),
			map[string]any{"action": "planner.create-goal", "planId": c.PlanID, "title": c.Title})
	}
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	resp, err := client.CallTool(ctx.Ctx, "CreateGoal", map[string]any{
		"planId": c.PlanID, "title": c.Title,
	})
	if err != nil {
		return fmt.Errorf("create goal: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Goal created", data)
}

type GoalsUpdateCmd struct {
	ID    string `arg:"" help:"Goal ID"`
	Title string `help:"New title" optional:""`
}

func (c *GoalsUpdateCmd) Run(ctx *commands.Context) error {
	if ctx.DryRun {
		return ctx.ValidateDryRun(plannerEndpoint(), "UpdateGoal", fmt.Sprintf("update goal %s", c.ID),
			map[string]any{"action": "planner.update-goal", "goalId": c.ID})
	}
	client := ctx.NewMCPClient(plannerEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	args := map[string]any{"goalId": c.ID}
	if c.Title != "" {
		args["title"] = c.Title
	}
	resp, err := client.CallTool(ctx.Ctx, "UpdateGoal", args)
	if err != nil {
		return fmt.Errorf("update goal: %w", err)
	}
	data, err := output.ExtractContent(resp)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Goal updated", data)
}
