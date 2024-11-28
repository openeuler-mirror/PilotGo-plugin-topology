/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Thu Nov 16 09:09:51 2023 +0800
 */
import * as echarts from 'echarts/core';
import {
  BarChart,
  // 系列类型的定义后缀都为 SeriesOption
  type BarSeriesOption,
  LineChart,
  type LineSeriesOption
} from 'echarts/charts';
import {
  TitleComponent,
  // 组件类型的定义后缀都为 ComponentOption
  type TitleComponentOption,
  TooltipComponent,
  type TooltipComponentOption,
  GridComponent,
  type GridComponentOption,
  // 数据集组件
  DatasetComponent,
  type DatasetComponentOption,
  // 内置数据转换器组件 (filter, sort)
  TransformComponent
} from 'echarts/components';
import { LabelLayout, UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type ECOption = echarts.ComposeOption<
  | BarSeriesOption
  | LineSeriesOption
  | TitleComponentOption
  | TooltipComponentOption
  | GridComponentOption
  | DatasetComponentOption
>;

// 注册必须的组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DatasetComponent,
  TransformComponent,
  BarChart,
  LineChart,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer
]);


export default echarts;