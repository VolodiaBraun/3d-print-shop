import type { DataProvider } from "@refinedev/core";
import api from "../lib/api";

interface CategoryNode {
  id: number;
  children?: CategoryNode[];
  [key: string]: unknown;
}

function flattenTree(nodes: CategoryNode[]): CategoryNode[] {
  const result: CategoryNode[] = [];
  for (const node of nodes) {
    result.push(node);
    if (node.children?.length) {
      result.push(...flattenTree(node.children));
    }
  }
  return result;
}

function findInTree(nodes: CategoryNode[], id: number): CategoryNode | null {
  for (const node of nodes) {
    if (node.id === id) return node;
    if (node.children?.length) {
      const found = findInTree(node.children, id);
      if (found) return found;
    }
  }
  return null;
}

export const dataProvider: DataProvider = {
  getApiUrl: () => "/api/v1",

  getList: async ({ resource, pagination, sorters, filters }) => {
    // Categories: tree endpoint, flatten for list
    if (resource === "categories") {
      const { data: resp } = await api.get("/categories");
      const tree: CategoryNode[] = resp.data;
      const flat = flattenTree(tree);
      return { data: flat, total: flat.length };
    }

    // Generic paginated resource (products, etc.)
    const params: Record<string, string | number> = {};
    if (pagination?.current) params.page = pagination.current;
    if (pagination?.pageSize) params.limit = pagination.pageSize;

    if (sorters?.length) {
      const s = sorters[0];
      params.sort = s.order === "desc" ? `-${s.field}` : s.field;
    }

    if (filters?.length) {
      for (const f of filters) {
        if ("field" in f && f.value !== undefined && f.value !== null && f.value !== "") {
          params[f.field] = f.value;
        }
      }
    }

    // Admin endpoints for listing (includes inactive items)
    const { data: resp } = await api.get(`/admin/${resource}`, { params });
    return {
      data: resp.data,
      total: resp.meta?.total ?? resp.data?.length ?? 0,
    };
  },

  getOne: async ({ resource, id }) => {
    if (resource === "categories") {
      const { data: resp } = await api.get("/categories");
      const found = findInTree(resp.data, Number(id));
      if (!found) throw new Error("Not found");
      return { data: found };
    }

    // Admin endpoint for getting by ID
    const { data: resp } = await api.get(`/admin/${resource}/${id}`);
    return { data: resp.data };
  },

  create: async ({ resource, variables }) => {
    const { data: resp } = await api.post(`/admin/${resource}`, variables);
    return { data: resp.data };
  },

  update: async ({ resource, id, variables }) => {
    const { data: resp } = await api.put(`/admin/${resource}/${id}`, variables);
    return { data: resp.data };
  },

  deleteOne: async ({ resource, id }) => {
    await api.delete(`/admin/${resource}/${id}`);
    return { data: { id } } as never;
  },
};
