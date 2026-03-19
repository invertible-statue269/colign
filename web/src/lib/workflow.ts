import { createClient } from "@connectrpc/connect";
import { WorkflowService } from "@/gen/proto/workflow/v1/workflow_pb";
import { transport } from "./connect";

export const workflowClient = createClient(WorkflowService, transport);
