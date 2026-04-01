"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
const pg_1 = require("pg");
const Y = __importStar(require("yjs"));
const html_to_yjs_1 = require("./html-to-yjs");
const prosemirror_1 = require("./prosemirror");
const dbUrl = new URL(process.env.DATABASE_URL ?? "postgres://postgres:postgres@localhost:5432/colign");
const searchPath = dbUrl.searchParams.get("search_path") ?? "public";
dbUrl.searchParams.delete("search_path");
const pool = new pg_1.Pool({ connectionString: dbUrl.toString() });
pool.on("connect", (client) => {
    client.query(`SET search_path TO ${searchPath}`);
});
async function main() {
    await migrateDocuments();
    await migrateDocumentVersions();
    await pool.end();
}
async function migrateDocuments() {
    const client = await pool.connect();
    try {
        await client.query("BEGIN");
        const result = await client.query("SELECT id, content FROM documents WHERE type <> 'proposal' ORDER BY id ASC");
        let migrated = 0;
        let skipped = 0;
        for (const row of result.rows) {
            if (!row.content || (0, prosemirror_1.isProseMirrorJSONContent)(row.content)) {
                skipped += 1;
                continue;
            }
            const nextContent = convertHTMLToProseMirrorJSON(row.content);
            await client.query("UPDATE documents SET content = $1 WHERE id = $2", [nextContent, row.id]);
            migrated += 1;
        }
        await client.query("COMMIT");
        console.log(`documents: migrated=${migrated} skipped=${skipped}`);
    }
    catch (err) {
        await client.query("ROLLBACK");
        throw err;
    }
    finally {
        client.release();
    }
}
async function migrateDocumentVersions() {
    const client = await pool.connect();
    try {
        await client.query("BEGIN");
        const result = await client.query(`SELECT dv.id, dv.content
       FROM document_versions AS dv
       JOIN documents AS d ON d.id = dv.document_id
       WHERE d.type <> 'proposal'
       ORDER BY dv.id ASC`);
        let migrated = 0;
        let skipped = 0;
        for (const row of result.rows) {
            if (!row.content || (0, prosemirror_1.isProseMirrorJSONContent)(row.content)) {
                skipped += 1;
                continue;
            }
            const nextContent = convertHTMLToProseMirrorJSON(row.content);
            await client.query("UPDATE document_versions SET content = $1 WHERE id = $2", [nextContent, row.id]);
            migrated += 1;
        }
        await client.query("COMMIT");
        console.log(`document_versions: migrated=${migrated} skipped=${skipped}`);
    }
    catch (err) {
        await client.query("ROLLBACK");
        throw err;
    }
    finally {
        client.release();
    }
}
function convertHTMLToProseMirrorJSON(content) {
    const doc = new Y.Doc();
    const fragment = doc.getXmlFragment("default");
    (0, html_to_yjs_1.htmlToYXmlFragment)(doc, fragment, content);
    return JSON.stringify((0, prosemirror_1.yXmlFragmentToProseMirrorJSON)(fragment));
}
void main().catch(async (err) => {
    console.error("document migration failed", err);
    await pool.end();
    process.exit(1);
});
