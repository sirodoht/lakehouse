import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import Collaboration from "@tiptap/extension-collaboration";
// import CollaborationCursor from '@tiptap/extension-collaboration-cursor'
import { HocuspocusProvider } from "@hocuspocus/provider";

// websocket provider
const provider = new HocuspocusProvider({
  url: "ws://127.0.0.1:8001",
  name: DOCUMENT_NAME, // eslint-disable-line no-undef
});

new Editor({
  element: document.querySelector("#id_body"),
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
