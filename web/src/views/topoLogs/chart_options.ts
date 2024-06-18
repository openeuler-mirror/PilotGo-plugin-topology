import * as echarts from 'echarts';

type EChartsOption = echarts.EChartsOption;
let bar_option: EChartsOption & {
  xAxis: {
    data?: any
  }
};

bar_option = {
  title: {
        text: '日志数量',
        left: '1%'
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
  xAxis: {},
  yAxis: {},
  dataZoom: [
    {
      start: 0,
      end: 50,
      bottom: 4,
      height: 18,
    },
    {
      type: 'inside'
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
  series: []
};

export default bar_option;
