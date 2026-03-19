"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { useState, useCallback, useEffect } from "react";
import { Button } from "@/components/ui/button";
import TurndownService from "turndown";
import { marked } from "marked";

const turndown = new TurndownService({
  headingStyle: "atx",
  bulletListMarker: "-",
  codeBlockStyle: "fenced",
});

interface SpecEditorProps {
  initialContent?: string;
  placeholder?: string;
  onSave?: (content: string) => void;
  readOnly?: boolean;
}

export function SpecEditor({
  initialContent = "",
  placeholder = "Start writing...",
  onSave,
  readOnly = false,
}: SpecEditorProps) {
  const [mode, setMode] = useState<"preview" | "edit" | "source">("preview");
  const [sourceContent, setSourceContent] = useState(initialContent);
  const [htmlContent, setHtmlContent] = useState(initialContent);
  const [saveStatus, setSaveStatus] = useState<"saved" | "saving" | "error" | "idle">("idle");

  const editor = useEditor({
    extensions: [StarterKit, Placeholder.configure({ placeholder })],
    content: initialContent,
    editable: !readOnly && mode === "edit",
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      const html = editor.getHTML();
      setHtmlContent(html);
      debouncedSave(html);
    },
  });

  const debouncedSave = useCallback(
    (() => {
      let timeout: NodeJS.Timeout;
      return (content: string) => {
        setSaveStatus("idle");
        clearTimeout(timeout);
        timeout = setTimeout(async () => {
          if (onSave) {
            setSaveStatus("saving");
            try {
              onSave(content);
              setSaveStatus("saved");
            } catch {
              setSaveStatus("error");
            }
          }
        }, 500);
      };
    })(),
    [onSave],
  );

  useEffect(() => {
    if (editor && initialContent) {
      editor.commands.setContent(initialContent);
      setHtmlContent(initialContent);
    }
  }, [initialContent, editor]);

  // Sync editable state when mode changes
  useEffect(() => {
    if (editor) {
      editor.setEditable(!readOnly && mode === "edit");
      if (mode === "edit") {
        editor.commands.focus();
      }
    }
  }, [mode, editor, readOnly]);

  function enterEdit() {
    setMode("edit");
  }

  function enterSource() {
    if (editor) {
      const md = turndown.turndown(editor.getHTML());
      setSourceContent(md);
    }
    setMode("source");
  }

  function exitToPreview() {
    if (mode === "source" && editor) {
      const html = marked.parse(sourceContent) as string;
      editor.commands.setContent(html);
      setHtmlContent(html);
      debouncedSave(html);
    }
    setMode("preview");
  }

  // Preview mode — rendered HTML, click to edit
  if (mode === "preview") {
    return (
      <div className="rounded-lg border border-border/50">
        <div className="flex items-center justify-between border-b border-border/50 px-4 py-2">
          <span className="text-xs text-muted-foreground">
            {saveStatus === "saved" && "Saved"}
            {saveStatus === "saving" && "Saving..."}
            {saveStatus === "error" && "Save failed"}
          </span>
          {!readOnly && (
            <Button
              variant="ghost"
              size="sm"
              onClick={enterEdit}
              className="cursor-pointer text-xs"
            >
              Edit
            </Button>
          )}
        </div>
        <div
          className="prose prose-invert max-w-none cursor-pointer p-6 min-h-[300px]"
          onClick={readOnly ? undefined : enterEdit}
          dangerouslySetInnerHTML={{
            __html: htmlContent || '<p class="text-muted-foreground">Click to start editing...</p>',
          }}
        />
      </div>
    );
  }

  // Edit mode — Tiptap WYSIWYG
  if (mode === "edit") {
    return (
      <div className="rounded-lg border border-primary/30">
        {/* Toolbar */}
        <div className="flex items-center justify-between border-b border-border/50 px-3 py-2">
          <div className="flex items-center gap-1">
            {editor && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => editor.chain().focus().toggleBold().run()}
                  className={`cursor-pointer text-xs ${editor.isActive("bold") ? "bg-accent" : ""}`}
                >
                  B
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => editor.chain().focus().toggleItalic().run()}
                  className={`cursor-pointer text-xs ${editor.isActive("italic") ? "bg-accent" : ""}`}
                >
                  I
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
                  className={`cursor-pointer text-xs ${editor.isActive("heading", { level: 2 }) ? "bg-accent" : ""}`}
                >
                  H2
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
                  className={`cursor-pointer text-xs ${editor.isActive("heading", { level: 3 }) ? "bg-accent" : ""}`}
                >
                  H3
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => editor.chain().focus().toggleBulletList().run()}
                  className={`cursor-pointer text-xs ${editor.isActive("bulletList") ? "bg-accent" : ""}`}
                >
                  List
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => editor.chain().focus().toggleCodeBlock().run()}
                  className={`cursor-pointer text-xs ${editor.isActive("codeBlock") ? "bg-accent" : ""}`}
                >
                  Code
                </Button>
              </>
            )}
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">
              {saveStatus === "saving" && "Saving..."}
              {saveStatus === "saved" && "Saved"}
              {saveStatus === "error" && "Save failed"}
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={enterSource}
              className="cursor-pointer text-xs text-muted-foreground"
            >
              Source
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={exitToPreview}
              className="cursor-pointer text-xs"
            >
              Done
            </Button>
          </div>
        </div>

        <div className="min-h-[400px] p-4">
          <EditorContent editor={editor} className="prose prose-invert max-w-none" />
        </div>
      </div>
    );
  }

  // Source mode — raw HTML
  return (
    <div className="rounded-lg border border-border/50">
      <div className="flex items-center justify-between border-b border-border/50 px-3 py-2">
        <span className="text-xs text-muted-foreground">Source</span>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setMode("edit")}
            className="cursor-pointer text-xs text-muted-foreground"
          >
            WYSIWYG
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={exitToPreview}
            className="cursor-pointer text-xs"
          >
            Done
          </Button>
        </div>
      </div>
      <div className="min-h-[400px] p-4">
        <textarea
          className="h-full min-h-[400px] w-full resize-none bg-transparent font-mono text-sm text-foreground outline-none"
          value={sourceContent}
          onChange={(e) => setSourceContent(e.target.value)}
        />
      </div>
    </div>
  );
}
