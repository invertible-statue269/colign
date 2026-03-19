"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Collaboration from "@tiptap/extension-collaboration";
import CollaborationCursor from "@tiptap/extension-collaboration-cursor";
import { HocuspocusProvider } from "@hocuspocus/provider";
import { useEffect, useMemo } from "react";
import * as Y from "yjs";

interface CollaborativeEditorProps {
  documentId: string;
  userName: string;
  userColor?: string;
}

export function CollaborativeEditor({
  documentId,
  userName,
  userColor = "#3b82f6",
}: CollaborativeEditorProps) {
  const ydoc = useMemo(() => new Y.Doc(), []);

  const provider = useMemo(
    () =>
      new HocuspocusProvider({
        url: process.env.NEXT_PUBLIC_HOCUSPOCUS_URL ?? "ws://localhost:1234",
        name: documentId,
        document: ydoc,
      }),
    [documentId, ydoc]
  );

  const editor = useEditor({
    extensions: [
      StarterKit.configure({ history: false }),
      Collaboration.configure({ document: ydoc }),
      CollaborationCursor.configure({
        provider,
        user: { name: userName, color: userColor },
      }),
    ],
    immediatelyRender: false,
  });

  useEffect(() => {
    return () => {
      provider.destroy();
      ydoc.destroy();
    };
  }, [provider, ydoc]);

  return (
    <div className="rounded-lg border">
      <div className="min-h-[400px] p-4">
        <EditorContent editor={editor} className="prose max-w-none" />
      </div>
    </div>
  );
}
