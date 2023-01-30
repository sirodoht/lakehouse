import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import Collaboration from "@tiptap/extension-collaboration";
import { HocuspocusProvider } from "@hocuspocus/provider";

// Set up the Hocuspocus WebSocket provider
const provider = new HocuspocusProvider({
  url: "ws://127.0.0.1:8001",
  name: "example-document",
});

new Editor({
  element: document.querySelector(".element"),
  content: "<p>Hello World!</p>",
  extensions: [
    StarterKit.configure({
      // The Collaboration extension comes with its own history handling
      history: false,
    }),
    // Register the document with Tiptap
    Collaboration.configure({
      document: provider.document,
    }),
  ],
});
