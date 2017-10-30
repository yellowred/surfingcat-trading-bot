<template>
  <div class="container-fluid">
    <div id="graph-container"></div>
    <button class="btn btn-primary" v-on:click="getQuote()">Update</button>
    <div id="graph-container2"></div>
    <button class="btn btn-primary" v-on:click="getQuote2()">Update</button>
  </div>
</template>


<script>
import anychart from 'anychart'
export default {
  data () {
    return {
      quote: ''
    }
  },
  methods: {
    getQuote () {
      let chartData = {chart: null, ema: null}
      this.$http.get('http://localhost:3026/chart/usdbtc')
      .then(res => {
        chartData.chart = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator?name=trima&market=USDT-BTC&interval=30') })
      .then(res => {
        chartData.ema50 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator?name=trima&market=USDT-BTC&interval=12') })
      .then(res => {
        chartData.ema20 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator?name=wma&market=USDT-BTC&interval=50') })
      .then(res => {
        chartData.wma50 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator?name=wma&market=USDT-BTC&interval=20') })
      .then(res => {
        chartData.wma20 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator?name=httrendline&market=USDT-BTC&interval=20') })
      .then(res => {
        chartData.trend = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => {
        return anychart.onDocumentReady(() => {
          var chart = anychart.stock(false)

          chart.title('BTC Chart')
          var plot = chart.plot()

          var dataTable = anychart.data.table()
          dataTable.addData(chartData.chart)
          var series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'USDT-BTC'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.ema50)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'EMA 50'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.ema20)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'EMA 20'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.wma50)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'WMA 50'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.wma20)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'WMA 20'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.trend)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'Trend'})

          chart.container('graph-container').draw()
        })
      })
    },
    getQuote2 () {
      let chartData = {chart: null, ema: null}

      this.$http.get('http://localhost:3026/strategy/test?market=USDT-BTC')
      .then(res => {
        chartData.testing = res.data.Balances.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then((res) => {
        return anychart.onDocumentReady(() => {
          var chart = anychart.stock(false)

          chart.title('Testing Chart')
          var plot = chart.plot()

          var dataTable = anychart.data.table()
          dataTable.addData(chartData.testing)
          var series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'TESTING'})

          chart.container('graph-container2').draw()
        })
      })
    }
  }
}
</script>
