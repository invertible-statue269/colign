import { createClient } from "@connectrpc/connect";
import { ProjectService } from "@/gen/proto/project/v1/project_pb";
import { transport } from "./connect";

export const projectClient = createClient(ProjectService, transport);
