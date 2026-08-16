/** Browser stub so the client bundle never ships linkedom. */
export function parseHTML(): never {
  throw new Error("linkedom is not available in the browser");
}

export const Node = class Node {};
export const NodeFilter = {};
export class DOMParser {}
