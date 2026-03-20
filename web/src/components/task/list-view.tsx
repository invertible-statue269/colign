"use client";

import { useI18n } from "@/lib/i18n";
import { TaskRow } from "./task-row";
import { InlineTaskInput } from "./inline-task-input";

interface ListViewProps {
  tasks: Array<{
    id: bigint;
    title: string;
    description: string;
    status: string;
    specRef: string;
    assigneeId?: bigint;
    assigneeName: string;
    orderIndex: number;
  }>;
  members: Array<{ userId: bigint; userName: string }>;
  onCreateTask: (title: string, status: string) => void;
  onUpdateTask: (id: bigint, fields: Record<string, unknown>) => void;
  onDeleteTask: (id: bigint) => void;
}

export function ListView({
  tasks,
  members,
  onCreateTask,
  onUpdateTask,
  onDeleteTask,
}: ListViewProps) {
  const { t } = useI18n();

  const groups = [
    { status: "todo", label: t("tasks.statusTodo"), color: "bg-gray-400" },
    { status: "in_progress", label: t("tasks.statusInProgress"), color: "bg-yellow-500" },
    { status: "done", label: t("tasks.statusDone"), color: "bg-green-500" },
  ];

  return (
    <div className="space-y-6">
      {groups.map((group) => {
        const groupTasks = tasks
          .filter((task) => task.status === group.status)
          .sort((a, b) => a.orderIndex - b.orderIndex);

        return (
          <div key={group.status}>
            {/* Header */}
            <div className="flex items-center gap-2 mb-2">
              <span className={`h-2 w-2 rounded-full ${group.color}`} />
              <span className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
                {group.label}
              </span>
              <span className="text-xs text-muted-foreground">{groupTasks.length}</span>
            </div>

            {/* Tasks */}
            <div className="space-y-1">
              {groupTasks.map((task) => (
                <TaskRow
                  key={String(task.id)}
                  task={task}
                  members={members}
                  onUpdate={onUpdateTask}
                  onDelete={onDeleteTask}
                />
              ))}
              <InlineTaskInput onSubmit={(title) => onCreateTask(title, group.status)} />
            </div>
          </div>
        );
      })}
    </div>
  );
}
