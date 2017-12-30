<template>
  <div class="container-fluid">
    <div class="row mt-3 mb-3">
      <h2>Bots</h2>
    </div>

    <div class="row">
      <div class="card p-0 m-2"  style="width: 22rem;" v-for="item in bots">
        <div class="card-header">
          <h4 v-bind:class="{ 'text-muted': item.Status == 'finished', 'card-title': true }"><span class="badge badge-success" v-if="item.Status == 'started'">started</span> {{item.Market}} {{ item.Uuid }}</h4>
        </div>
        <div class="card-body">
          
          <h6 class="card-subtitle">{{item.Started}} &mdash; {{item.Finished}}</h6>

          <small>
          <table class="table table-bordered mt-2 mb-2 table-sm">
            <tr v-for="(val, key) in JSON.parse(item.Config)">
              <td>{{ key }}</td>
              <td>{{ val }}</td>
            </tr>
          </table>
          </small>

          <button v-if="item.Actions != ''" class="btn btn-light" type="button" data-toggle="collapse" v-bind:data-target="'#ActionsList_' + item.Uuid">
            actions list
          </button>

          <table class="collapse table table-striped mt-2 mb-2 table-sm table-responsive" v-if="item.Actions != ''" v-bind:id="'ActionsList_' + item.Uuid">
            
            <tr>
              <th>Act</th>
              <th>Market</th>
              <th>Amount</th>
              <th>Rate</th>
            </tr>

            <tbody>
            <tr v-for="action in item.Actions"><td v-for="(val, key) in action.split(',')">

              <span class="badge badge-success" v-if="val == 'market_buy'">buy</span>
              <span class="badge badge-danger" v-else-if="val == 'market_sell'">sell</span>
              <span style="font-size:0.7em" v-else-if="key == 1">{{ val }}</span>
              <span v-else>{{ val }}</span>
              
              
            </td></tr>
            </tbody>
          </table>
        </div>
      </div>
      
    </div>
  </div>
</template>


<script>
import auth from '../auth'

export default {
  data () {
    return {
      bots: []
    }
  },
  mounted () {
    var self = this
    this.$http.get('http://localhost:3026/api/trader/status', {headers: auth.getAuthHeader()})
      .then(res => {
        self.bots = res.body.filter(item => item.Config !== '')
      }, res => {
        if (res.status === 401) {
          auth.logout(this)
          self.$router.replace('/login')
        }
      })
  },
  methods: {
  }
}
</script>
