import { createClient } from "@connectrpc/connect";
import { ApiTokenService } from "@/gen/proto/apitoken/v1/apitoken_pb";
import { transport } from "./connect";

export const apiTokenClient = createClient(ApiTokenService, transport);
