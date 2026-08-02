package planner

import (
	"fmt"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

// PlannerCmd groups all Planner subcommands.
type PlannerCmd struct {
	Plans PlansCmd `cmd:"" help:"Planner plans"`
	Tasks TasksCmd `cmd:"" help:"Planner tasks"`
	Goals GoalsCmd `cmd:"" help:"Planner goals"`
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
	data, err := ctx.CallToolData(plannerEndpoint(), "QueryPlans", "list plans", map[string]any{})
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("plans", output.PlannerPlanColumns, data, c.Max, "plans", "value")
}

type PlansGetCmd struct {
	ID string `arg:"" help:"Plan ID"`
}

func (c *PlansGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(plannerEndpoint(), "GetPlan", "get plan", map[string]any{"planId": c.ID})
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
	data, err := ctx.CallToolData(plannerEndpoint(), "CreatePlan", "create plan", map[string]any{"title": c.Title})
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

	args := map[string]any{"planId": c.ID}
	if c.Title != "" {
		args["title"] = c.Title
	}
	data, err := ctx.CallToolData(plannerEndpoint(), "UpdatePlan", "update plan", args)
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
	data, err := ctx.CallToolData(plannerEndpoint(), "QueryTasksInPlan", "list tasks", map[string]any{"planId": c.PlanID})
	if err != nil {
		return err
	}
	return ctx.Output.PrintListFromData("tasks", output.PlannerTaskColumns, data, c.Max, "tasks", "value")
}

type TasksGetCmd struct {
	ID string `arg:"" help:"Task ID"`
}

func (c *TasksGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(plannerEndpoint(), "GetTask", "get task", map[string]any{"taskId": c.ID})
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
	data, err := ctx.CallToolData(plannerEndpoint(), "CreateTask", "create task", map[string]any{
		"planId": c.PlanID, "title": c.Title,
	})
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

	args := map[string]any{"taskId": c.ID}
	if c.Title != "" {
		args["title"] = c.Title
	}
	data, err := ctx.CallToolData(plannerEndpoint(), "UpdateTask", "update task", args)
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
	data, err := ctx.CallToolData(plannerEndpoint(), "QueryGoalsInPlan", "list goals", map[string]any{"planId": c.PlanID})
	if err != nil {
		return err
	}
	return ctx.Output.PrintItem(data)
}

type GoalsGetCmd struct {
	ID string `arg:"" help:"Goal ID"`
}

func (c *GoalsGetCmd) Run(ctx *commands.Context) error {
	data, err := ctx.CallToolData(plannerEndpoint(), "GetGoal", "get goal", map[string]any{"goalId": c.ID})
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
	data, err := ctx.CallToolData(plannerEndpoint(), "CreateGoal", "create goal", map[string]any{
		"planId": c.PlanID, "title": c.Title,
	})
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

	args := map[string]any{"goalId": c.ID}
	if c.Title != "" {
		args["title"] = c.Title
	}
	data, err := ctx.CallToolData(plannerEndpoint(), "UpdateGoal", "update goal", args)
	if err != nil {
		return err
	}
	return ctx.Output.PrintMutation("Goal updated", data)
}
