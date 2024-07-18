
type ID = {
  id: string;
}
interface UserGraphData {
  nodes: {
    id: string,
    parentId: string,
    data: object,
  }[],
  edges: {
    id: string,
    source: string,
    target: string,
    data: object
  }[],
  combos: {
    id: string,
    parentId: string,
    data: object,
  }[]
}

// TreeGraph
type TreeGraphData = {
  id: string;
  [key: string]: unknown;
  children: TreeGraphData[];
}

type GraphSpec = {
  data: {
    type: 'fetch' | 'tree' | 'graph';
    value: 'string' | TreeGraphData | GraphData;
    roots: ID[]; // 在 type 为 graph 时需要
  },
  // ... 其他配置项
}


// Graph
type GraphData = {
  nodes: {
    id: string;
    parentId: string;
    // 若使用 GraphData 作为树图数据，需要指定 treeParentId，
    // G6 将检查由此是否能够构造树，如果不可以，将尽量兼容并打印警告
    // 若所有节点均未指定 treeParentId，则将使用最小生成树算法构造树结构
    treeParentId: string;
    [key: string]: unknown;
  }[];
  edges: {
    id: string;
    source: string;
    target: string;
    [key: string]: unknown;
  }[],
  combos: {
    id: string;
    parentId: string;
    [key: string]: unknown;
  }[]
}

export interface Config {
  batchId: number;
  id: number;
  conf_name: string;
  create_time: string;
  update_time?: string;
  description: string;
  [key: string]: unknown;
}

// *接口api返回结果约束不含data
export interface Result {
  code: number;
  msg?: string;
}

// *接口api返回结果含有page信息
export interface ResultData extends Result {
  data: Config[];
  ok?: Boolean;
  page: number;
  size: number;
  total: number;
}

// topo
export interface TopoCustomFormType {
  conf_name: string;
  conf_time: string;
  batchId: number;
  node_rules: [[{ rule_condition: any, rule_type: string }, { rule_condition: any, rule_type: string }]];
  description: string;
  [key: string]: unknown;
}

export type logData = {
  name: string,
  data: (string | number)[][]
};