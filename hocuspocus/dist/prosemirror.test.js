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
const Y = __importStar(require("yjs"));
const html_to_yjs_1 = require("./html-to-yjs");
const prosemirror_1 = require("./prosemirror");
function toFragment(content) {
    const doc = new Y.Doc();
    const fragment = doc.getXmlFragment("default");
    (0, html_to_yjs_1.htmlToYXmlFragment)(doc, fragment, content);
    return fragment;
}
{
    const fragment = toFragment("<h2>Design</h2><p>Hello <code>world()</code></p><pre><code class=\"language-go\">fmt.Println(1)</code></pre>");
    const json = (0, prosemirror_1.yXmlFragmentToProseMirrorJSON)(fragment);
    console.assert(json.type === "doc", "expected doc root");
    console.assert(json.content?.[0]?.type === "heading", "expected heading node");
    console.assert(json.content?.[1]?.type === "paragraph", "expected paragraph node");
    console.assert(json.content?.[2]?.type === "codeBlock", "expected codeBlock node");
    console.log("PASS: Y.js -> ProseMirror JSON");
}
{
    const doc = new Y.Doc();
    const fragment = doc.getXmlFragment("default");
    const content = {
        type: "doc",
        content: [
            { type: "heading", attrs: { level: 2 }, content: [{ type: "text", text: "API" }] },
            {
                type: "paragraph",
                content: [
                    { type: "text", text: "Use " },
                    { type: "text", text: "client()", marks: [{ type: "code" }] },
                ],
            },
        ],
    };
    (0, prosemirror_1.proseMirrorJSONToYXmlFragment)(doc, fragment, content);
    console.assert(fragment.length === 2, `expected 2 top-level nodes, got ${fragment.length}`);
    const paragraph = fragment.get(1);
    const text = paragraph.get(0);
    console.assert(text.toDelta().some((op) => typeof op.insert === "string" && op.attributes?.code), "expected code mark to be preserved");
    console.log("PASS: ProseMirror JSON -> Y.js");
}
