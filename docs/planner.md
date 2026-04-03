# Planner

Manage Microsoft Planner plans, tasks, and goals.

## Commands

### Plans

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `planner plans list` | List plans | `--max` |
| `planner plans get` | Get a plan by ID | `<plan-id>` |
| `planner plans create` | Create a plan | `<title>` |
| `planner plans update` | Update a plan | `<plan-id>` `--title` |

### Tasks

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `planner tasks list` | List tasks in a plan | `<plan-id>` `--max` |
| `planner tasks get` | Get a task by ID | `<task-id>` |
| `planner tasks create` | Create a task | `<plan-id>` `<title>` |
| `planner tasks update` | Update a task | `<task-id>` `--title` |

### Goals

| Command | Description | Key Arguments |
|---------|-------------|---------------|
| `planner goals list` | List goals in a plan | `<plan-id>` |
| `planner goals get` | Get a goal by ID | `<goal-id>` |
| `planner goals create` | Create a goal | `<plan-id>` `<title>` |
| `planner goals update` | Update a goal | `<goal-id>` `--title` |

## Examples

```bash
# List all your plans
a365 planner plans list

# Create a new plan and inspect it
a365 planner plans create "Q3 Roadmap"
a365 planner plans get PLAN_ID

# Rename a plan
a365 planner plans update PLAN_ID --title "Q3 Roadmap (revised)"

# List tasks in a plan
a365 planner tasks list PLAN_ID --max 20

# Create and update tasks
a365 planner tasks create PLAN_ID "Write design doc"
a365 planner tasks update TASK_ID --title "Write design doc (v2)"

# Get task details
a365 planner tasks get TASK_ID

# Manage goals for a plan
a365 planner goals list PLAN_ID
a365 planner goals create PLAN_ID "Ship MVP by end of sprint"
a365 planner goals update GOAL_ID --title "Ship MVP by July 31"
a365 planner goals get GOAL_ID
```
