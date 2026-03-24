import { createClient } from "@connectrpc/connect";
import { AIConfigService } from "@/gen/proto/aiconfig/v1/aiconfig_pb";
import { transport } from "./connect";

export const aiConfigClient = createClient(AIConfigService, transport);
