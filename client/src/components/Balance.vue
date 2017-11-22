<template>
  <div class="col-sm-6 col-sm-offset-3">
    <h1>Balance</h1>
    <button class="btn btn-warning" v-on:click="getQuote()">Get balances</button>
    <table class="balances-area" v-if="balances">
      <tr v-for="item in balances"><td>{{ item.Currency }} </td><td>{{ item.Balance }}</td></tr>
    </table>
  </div>
</template>

<script>
export default {
  data () {
    return {
      balances: ''
    }
  },
  methods: {
    getQuote () {
      this.$http.get('http://localhost:3026/trader/balance')
      .then(response => {
        this.balances = response.body
      }, response => {
        console.log(response)
      })
    }
  }
}
</script>
