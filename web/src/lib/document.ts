import { createClient } from "@connectrpc/connect";
import { DocumentService } from "@/gen/proto/document/v1/document_pb";
import { transport } from "./connect";

export const documentClient = createClient(DocumentService, transport);
