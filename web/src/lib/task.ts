import { createClient } from "@connectrpc/connect";
import { TaskService } from "@/gen/proto/task/v1/task_pb";
import { transport } from "./connect";

export const taskClient = createClient(TaskService, transport);
