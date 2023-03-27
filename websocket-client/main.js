import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import Collaboration from "@tiptap/extension-collaboration";
// import CollaborationCursor from '@tiptap/extension-collaboration-cursor'
import * as Y from 'yjs'
import { HocuspocusProvider } from "@hocuspocus/provider";


console.log("DOCUMENT_NAME:", DOCUMENT_NAME);
console.log("DOCUMENT_BODY:", DOCUMENT_BODY);

// websocket provider
const provider = new HocuspocusProvider({
  url: "ws://127.0.0.1:8001",
  name: DOCUMENT_NAME,
});

// instantiate tiptap editor
new Editor({
  // element: document.querySelector("#id_body"),
  element: document.querySelector(".element"),
  // content: DOCUMENT_BODY,
  extensions: [
    StarterKit.configure({
      // The Collaboration extension comes with its own history handling
      history: false,
    }),
    // Register the document with Tiptap
    Collaboration.configure({
      document: provider.document,
    }),
    // CollaborationCursor.configure({
    //   provider: provider,
    //   user: {
    //     name: "Cyndi Lauper",
    //     color: "#f783ac",
    //   },
    // }),
  ],
});
