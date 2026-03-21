import * as Y from "yjs";
import { htmlToYXmlFragment } from "./html-to-yjs";

function convert(html: string): Y.Doc {
  const doc = new Y.Doc();
  const fragment = doc.getXmlFragment("default");
  htmlToYXmlFragment(doc, fragment, html);
  return doc;
}

function fragmentToString(fragment: Y.XmlFragment): string {
  const items: string[] = [];
  fragment.forEach((item) => {
    if (item instanceof Y.XmlElement) {
      items.push(`${item.nodeName}(${JSON.stringify(item.getAttributes())})[${elementChildren(item)}]`);
    } else if (item instanceof Y.XmlText) {
      items.push(`text("${item.toString()}")`);
    }
  });
  return items.join(", ");
}

function elementChildren(el: Y.XmlElement): string {
  const items: string[] = [];
  el.forEach((child) => {
    if (child instanceof Y.XmlElement) {
      items.push(`${child.nodeName}(${JSON.stringify(child.getAttributes())})[${elementChildren(child)}]`);
    } else if (child instanceof Y.XmlText) {
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

  const h1 = fragment.get(0) as Y.XmlElement;
  console.assert(h1.nodeName === "heading", `Expected heading, got ${h1.nodeName}`);
  console.assert(h1.getAttribute("level") === "1", `Expected level 1, got ${h1.getAttribute("level")}`);

  const h2 = fragment.get(1) as Y.XmlElement;
  console.assert(h2.getAttribute("level") === "2", `Expected level 2`);

  const h3 = fragment.get(2) as Y.XmlElement;
  console.assert(h3.getAttribute("level") === "3", `Expected level 3`);

  console.log("PASS: headings");
}

// Test paragraph conversion
{
  const doc = convert("<p>Hello world</p><p>Second paragraph</p>");
  const fragment = doc.getXmlFragment("default");

  console.assert(fragment.length === 2, `Expected 2 paragraphs, got ${fragment.length}`);

  const p1 = fragment.get(0) as Y.XmlElement;
  console.assert(p1.nodeName === "paragraph", `Expected paragraph, got ${p1.nodeName}`);

  console.log("PASS: paragraphs");
}

// Test bullet list merging
{
  // markdownToHTML produces separate <ul> for each item, but they should be merged
  const doc = convert("<ul><li>Item 1</li></ul><ul><li>Item 2</li></ul><ul><li>Item 3</li></ul>");
  const fragment = doc.getXmlFragment("default");

  console.assert(fragment.length === 1, `Expected 1 bulletList (merged), got ${fragment.length}`);

  const list = fragment.get(0) as Y.XmlElement;
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

  const el0 = fragment.get(0) as Y.XmlElement;
  const el1 = fragment.get(1) as Y.XmlElement;
  const el2 = fragment.get(2) as Y.XmlElement;
  const el3 = fragment.get(3) as Y.XmlElement;

  console.assert(el0.nodeName === "heading", `[0] Expected heading`);
  console.assert(el1.nodeName === "paragraph", `[1] Expected paragraph`);
  console.assert(el2.nodeName === "bulletList", `[2] Expected bulletList`);
  console.assert(el3.nodeName === "paragraph", `[3] Expected paragraph`);

  console.log("PASS: mixed content");
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

  const p = fragment.get(0) as Y.XmlElement;
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

  const list = fragment.get(0) as Y.XmlElement;
  const listItem = list.get(0) as Y.XmlElement;
  console.assert(listItem.nodeName === "listItem", `Expected listItem`);

  const paragraph = listItem.get(0) as Y.XmlElement;
  console.assert(paragraph.nodeName === "paragraph", `Expected paragraph inside listItem, got ${paragraph.nodeName}`);

  console.log("PASS: listItem contains paragraph");
}

console.log("\nAll tests passed!");
