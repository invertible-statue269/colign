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
function convert(html) {
    const doc = new Y.Doc();
    const fragment = doc.getXmlFragment("default");
    (0, html_to_yjs_1.htmlToYXmlFragment)(doc, fragment, html);
    return doc;
}
function fragmentToString(fragment) {
    const items = [];
    fragment.forEach((item) => {
        if (item instanceof Y.XmlElement) {
            items.push(`${item.nodeName}(${JSON.stringify(item.getAttributes())})[${elementChildren(item)}]`);
        }
        else if (item instanceof Y.XmlText) {
            items.push(`text("${item.toString()}")`);
        }
    });
    return items.join(", ");
}
function elementChildren(el) {
    const items = [];
    el.forEach((child) => {
        if (child instanceof Y.XmlElement) {
            items.push(`${child.nodeName}(${JSON.stringify(child.getAttributes())})[${elementChildren(child)}]`);
        }
        else if (child instanceof Y.XmlText) {
            items.push(`text("${child.toString()}")`);
        }
    });
    return items.join(", ");
}
// Test heading conversion
{
    const doc = convert("<h1>Title</h1><h2>Subtitle</h2><h3>Section</h3>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 3, `Expected 3 elements, got ${fragment.length}`);
    const h1 = fragment.get(0);
    console.assert(h1.nodeName === "heading", `Expected heading, got ${h1.nodeName}`);
    console.assert(h1.getAttribute("level") === "1", `Expected level 1, got ${h1.getAttribute("level")}`);
    const h2 = fragment.get(1);
    console.assert(h2.getAttribute("level") === "2", `Expected level 2`);
    const h3 = fragment.get(2);
    console.assert(h3.getAttribute("level") === "3", `Expected level 3`);
    console.log("PASS: headings");
}
// Test paragraph conversion
{
    const doc = convert("<p>Hello world</p><p>Second paragraph</p>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 2, `Expected 2 paragraphs, got ${fragment.length}`);
    const p1 = fragment.get(0);
    console.assert(p1.nodeName === "paragraph", `Expected paragraph, got ${p1.nodeName}`);
    console.log("PASS: paragraphs");
}
// Test bullet list merging
{
    // markdownToHTML produces separate <ul> for each item, but they should be merged
    const doc = convert("<ul><li>Item 1</li></ul><ul><li>Item 2</li></ul><ul><li>Item 3</li></ul>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 1, `Expected 1 bulletList (merged), got ${fragment.length}`);
    const list = fragment.get(0);
    console.assert(list.nodeName === "bulletList", `Expected bulletList, got ${list.nodeName}`);
    let listItemCount = 0;
    list.forEach(() => listItemCount++);
    console.assert(listItemCount === 3, `Expected 3 list items, got ${listItemCount}`);
    console.log("PASS: bullet list merging");
}
// Test mixed content
{
    const doc = convert("<h2>Design</h2><p>Overview text</p><ul><li>Step 1</li><li>Step 2</li></ul><p>Conclusion</p>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 4, `Expected 4 elements (h2, p, ul, p), got ${fragment.length}`);
    const el0 = fragment.get(0);
    const el1 = fragment.get(1);
    const el2 = fragment.get(2);
    const el3 = fragment.get(3);
    console.assert(el0.nodeName === "heading", `[0] Expected heading`);
    console.assert(el1.nodeName === "paragraph", `[1] Expected paragraph`);
    console.assert(el2.nodeName === "bulletList", `[2] Expected bulletList`);
    console.assert(el3.nodeName === "paragraph", `[3] Expected paragraph`);
    console.log("PASS: mixed content");
}
// Test code block conversion
{
    const doc = convert('<h2>Example</h2><pre><code class="language-tsx">const x = 1;\nconsole.log(x);</code></pre>');
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 2, `Expected 2 elements (h2, codeBlock), got ${fragment.length}`);
    const codeBlock = fragment.get(1);
    console.assert(codeBlock.nodeName === "codeBlock", `Expected codeBlock, got ${codeBlock.nodeName}`);
    const codeText = codeBlock.get(0);
    console.assert(codeText.toString() === "const x = 1;\nconsole.log(x);", `Expected code block text to round-trip, got ${JSON.stringify(codeText.toString())}`);
    console.log("PASS: code blocks");
}
// Test inline formatting preserved in heading and list item
{
    const doc = convert("<h2>Hello <code>world()</code></h2><ol><li><strong>Step</strong> <code>one()</code></li></ol>");
    const fragment = doc.getXmlFragment("default");
    const heading = fragment.get(0);
    const list = fragment.get(1);
    const listItemParagraph = list.get(0).get(0);
    const headingText = heading.get(0);
    const listItemText = listItemParagraph.get(0);
    const headingDelta = headingText.toDelta();
    const listDelta = listItemText.toDelta();
    console.assert(headingDelta.some((op) => typeof op.insert === "string" && op.attributes?.code), "Expected heading inline code formatting to be preserved");
    console.assert(listDelta.some((op) => typeof op.insert === "string" && op.attributes?.bold), "Expected list item bold formatting to be preserved");
    console.assert(listDelta.some((op) => typeof op.insert === "string" && op.attributes?.code), "Expected list item inline code formatting to be preserved");
    console.log("PASS: inline formatting");
}
// Test blockquote conversion
{
    const doc = convert("<blockquote><p>This is a quote</p></blockquote>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 1, `Expected 1 blockquote, got ${fragment.length}`);
    const bq = fragment.get(0);
    console.assert(bq.nodeName === "blockquote", `Expected blockquote, got ${bq.nodeName}`);
    const p = bq.get(0);
    console.assert(p.nodeName === "paragraph", `Expected paragraph inside blockquote, got ${p.nodeName}`);
    const text = p.get(0);
    console.assert(text.toString() === "This is a quote", `Expected 'This is a quote', got '${text.toString()}'`);
    console.log("PASS: blockquote");
}
// Test blockquote with multiple paragraphs
{
    const doc = convert("<blockquote><p>First line</p><p>Second line</p></blockquote>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 1, `Expected 1 blockquote, got ${fragment.length}`);
    const bq = fragment.get(0);
    let childCount = 0;
    bq.forEach(() => childCount++);
    console.assert(childCount === 2, `Expected 2 paragraphs inside blockquote, got ${childCount}`);
    console.log("PASS: blockquote with multiple paragraphs");
}
// Test blockquote without <p> wrapper
{
    const doc = convert("<blockquote>Plain quoted text</blockquote>");
    const fragment = doc.getXmlFragment("default");
    const bq = fragment.get(0);
    console.assert(bq.nodeName === "blockquote", `Expected blockquote`);
    const p = bq.get(0);
    console.assert(p.nodeName === "paragraph", `Expected paragraph inside blockquote`);
    console.log("PASS: blockquote without p wrapper");
}
// Test blockquote with real marked.parse() output (contains newlines)
{
    const doc = convert("<blockquote>\n<p>Quoted via marked</p>\n</blockquote>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 1, `Expected 1 blockquote, got ${fragment.length}`);
    const bq = fragment.get(0);
    console.assert(bq.nodeName === "blockquote", `Expected blockquote, got ${bq.nodeName}`);
    const p = bq.get(0);
    console.assert(p.nodeName === "paragraph", `Expected paragraph inside blockquote`);
    const text = p.get(0);
    console.assert(text.toString() === "Quoted via marked", `Expected 'Quoted via marked', got '${text.toString()}'`);
    console.log("PASS: blockquote with marked.parse() newlines");
}
// Test blockquote with inline formatting
{
    const doc = convert("<blockquote><p>This is <strong>important</strong> text</p></blockquote>");
    const fragment = doc.getXmlFragment("default");
    const bq = fragment.get(0);
    const p = bq.get(0);
    const text = p.get(0);
    const delta = text.toDelta();
    console.assert(delta.some((op) => op.attributes?.bold && op.insert === "important"), "Expected bold formatting preserved in blockquote");
    console.assert(delta.some((op) => !op.attributes?.bold && op.insert.includes("text")), "Expected plain text after bold to NOT inherit bold");
    console.log("PASS: blockquote with inline formatting");
}
// Test blockquote in mixed content
{
    const doc = convert("<h2>Title</h2><blockquote><p>A quote</p></blockquote><p>After quote</p>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 3, `Expected 3 elements (h2, blockquote, p), got ${fragment.length}`);
    const el0 = fragment.get(0);
    const el1 = fragment.get(1);
    const el2 = fragment.get(2);
    console.assert(el0.nodeName === "heading", `[0] Expected heading, got ${el0.nodeName}`);
    console.assert(el1.nodeName === "blockquote", `[1] Expected blockquote, got ${el1.nodeName}`);
    console.assert(el2.nodeName === "paragraph", `[2] Expected paragraph, got ${el2.nodeName}`);
    console.log("PASS: blockquote in mixed content");
}
// Test empty input
{
    const doc = convert("");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 0, `Expected 0 elements for empty input`);
    console.log("PASS: empty input");
}
// Test HTML entity unescaping
{
    const doc = convert("<p>A &amp; B &lt; C</p>");
    const fragment = doc.getXmlFragment("default");
    const p = fragment.get(0);
    let textContent = "";
    p.forEach((child) => {
        if (child instanceof Y.XmlText) {
            textContent = child.toString();
        }
    });
    console.assert(textContent === "A & B < C", `Expected 'A & B < C', got '${textContent}'`);
    console.log("PASS: HTML entity unescaping");
}
// Test listItem has paragraph child (TipTap requirement)
{
    const doc = convert("<ul><li>Item</li></ul>");
    const fragment = doc.getXmlFragment("default");
    const list = fragment.get(0);
    const listItem = list.get(0);
    console.assert(listItem.nodeName === "listItem", `Expected listItem`);
    const paragraph = listItem.get(0);
    console.assert(paragraph.nodeName === "paragraph", `Expected paragraph inside listItem, got ${paragraph.nodeName}`);
    console.log("PASS: listItem contains paragraph");
}
// Test table conversion
{
    const doc = convert("<table><thead><tr><th>Stage</th><th>Owner</th></tr></thead><tbody><tr><td>Design</td><td>Ben</td></tr></tbody></table>");
    const fragment = doc.getXmlFragment("default");
    console.assert(fragment.length === 1, `Expected 1 table, got ${fragment.length}`);
    const table = fragment.get(0);
    console.assert(table.nodeName === "table", `Expected table, got ${table.nodeName}`);
    const headerRow = table.get(0);
    const bodyRow = table.get(1);
    console.assert(headerRow.nodeName === "tableRow", `Expected tableRow, got ${headerRow.nodeName}`);
    console.assert(bodyRow.nodeName === "tableRow", `Expected tableRow, got ${bodyRow.nodeName}`);
    const headerCell = headerRow.get(0);
    const bodyCell = bodyRow.get(0);
    console.assert(headerCell.nodeName === "tableHeader", `Expected tableHeader, got ${headerCell.nodeName}`);
    console.assert(bodyCell.nodeName === "tableCell", `Expected tableCell, got ${bodyCell.nodeName}`);
    const headerParagraph = headerCell.get(0);
    const bodyParagraph = bodyCell.get(0);
    console.assert(headerParagraph.nodeName === "paragraph", `Expected paragraph inside header cell`);
    console.assert(bodyParagraph.nodeName === "paragraph", `Expected paragraph inside body cell`);
    const headerText = headerParagraph.get(0);
    const bodyText = bodyParagraph.get(0);
    console.assert(headerText.toString() === "Stage", `Expected 'Stage', got '${headerText.toString()}'`);
    console.assert(bodyText.toString() === "Design", `Expected 'Design', got '${bodyText.toString()}'`);
    console.log("PASS: table conversion");
}
console.log("\nAll tests passed!");
