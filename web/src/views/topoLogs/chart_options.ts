import * as echarts from 'echarts';

type EChartsOption = echarts.EChartsOption;
let bar_option: EChartsOption & {
  xAxis: {
    data?: any[];
    type?: string;
    boundaryGap?: boolean;
    axisLabel?: unknown;
  },
  title: {
    text?: string;
  }
};

let colors = ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de', '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc'];

bar_option = {
  title: {
        text: '集群',
        left: '0'
      },
  tooltip: {
    confine: true,
    trigger: 'axis',
    axisPointer: {
      type: 'shadow'
    }
  },
  legend: {

  },
  grid: {
    left: '0',
    right: '2%',
    bottom: '8%',
    containLabel: true
  },
  xAxis: {
    type: 'time',
    boundaryGap:false
  },
  yAxis: {
    type:'value'
  },
  dataZoom: [
    {
      start: 0,
      end: 100,
      bottom: 4,
      height: 18,
    },
    {
      type: 'inside',
      start: 0,
      end: 10,
    }
  ],
  /* visualMap: {
    top: 0,
    right: 0,
    pieces: [
      {
        gt: 0,
        lte: 50,
        color: '#93CE07'
      },
      {
        gt: 50,
        lte: 100,
        color: '#FBDB0F'
      },
      {
        gt: 100,
        lte: 250,
        color: '#FC7D02'
      },
      {
        gt: 250,
        lte: 499,
        color: '#FD0100'
      }
    ],
    outOfRange: {
      color: '#999'
    }
  }, */
  series: [
    {
      name: "类别1",
      stack: "A",
      type: "bar",
      data: [],
    },
    {
      name: "类别2",
      stack: "A",
      type: "bar",
      data: [],
    }
  ]
};

export default bar_option;
