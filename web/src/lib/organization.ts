import { createClient } from "@connectrpc/connect";
import { OrganizationService } from "@/gen/proto/organization/v1/organization_pb";
import { transport } from "./connect";

export const orgClient = createClient(OrganizationService, transport);
