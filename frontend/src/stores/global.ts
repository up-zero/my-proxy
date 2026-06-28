import config from "@/config";
import { defineStore } from "pinia";

export interface CurrentNode {
  uuid: string;
  name: string;
  isLocal: boolean;
}

const STORAGE_KEY = `${config.name}:currentNode`;

export default defineStore("global", {
  state: () => ({
    currentNode: loadCurrentNode(),
  }),
  getters: {
    currentUuid: (state) => state.currentNode.uuid,
    isLocalNode: (state) => state.currentNode.isLocal,
  },
  actions: {
    setCurrentNode(node: CurrentNode) {
      this.currentNode = { ...node };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(node));
    },
    getCurrentNode(): CurrentNode {
      return { ...this.currentNode };
    },
  },
});

function loadCurrentNode(): CurrentNode {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as CurrentNode;
      if (parsed.uuid && parsed.name) {
        return parsed;
      }
    }
  } catch {
    // ignore
  }
  return {
    uuid: "node-local",
    name: "Local",
    isLocal: true,
  };
}
