<template>
  <div class="container-fluid">
    <div id="graph-container"></div>
    <button class="btn btn-primary" v-on:click="TestbedData()">Update</button>
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
    TestbedData () {
      let chartData = {chart: null}
      this.$http.get('http://localhost:3026/chart/testbed')
      .then(res => {
        chartData.chart = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator/testbed?name=trima&market=USDT-BTC&interval=30') })
      .then(res => {
        chartData.indi1 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator/testbed?name=trima&market=USDT-BTC&interval=12') })
      .then(res => {
        chartData.indi2 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator/testbed?name=wma&market=USDT-BTC&interval=50') })
      .then(res => {
        chartData.indi3 = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator/testbed?name=wma&market=USDT-BTC&interval=30') })
      .then(res => {
        chartData.indi4 = res.data.map(item => {
          return [item.Date, item.Value]
        })
        console.log('WMA', chartData.indi4)
      })
      .then(() => { return this.$http.get('http://localhost:3026/indicator/testbed?name=httrendline&market=USDT-BTC&interval=20') })
      .then(res => {
        chartData.trend = res.data.map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => { return this.$http.get('http://localhost:3026/strategy/test?market=USDT-BTC&strategy=dip') })
      .then(res => {
        chartData.buys = res.data.Actions.filter(item => item.Action === 0).map(item => {
          return [item.Date, item.Value]
        })
        chartData.sells = res.data.Actions.filter(item => item.Action === 1).map(item => {
          return [item.Date, item.Value]
        })
      })
      .then(() => {
        return anychart.onDocumentReady(() => {
          var chart = anychart.stock(false)

          chart.title('Testbed Data')
          var plot = chart.plot()

          var dataTable = anychart.data.table()
          dataTable.addData(chartData.chart)
          var series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'USDT-BTC'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.indi1)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'TRIMA 30'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.indi2)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'TRIMA 12'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.indi3)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'WMA 50'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.indi4)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'WMA 20'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.indi5)
          series = plot.line(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'Trend'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.buys)
          series = plot.marker(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'Buys'})

          dataTable = anychart.data.table()
          dataTable.addData(chartData.sells)
          series = plot.marker(dataTable.mapAs({'value': 1}))
          series.legendItem({text: 'Sells'})

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
