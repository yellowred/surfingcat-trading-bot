<template>
  <div class="container-fluid">
    <div class="row">
      
      <div class="col-lg-4">
        <h3>Platform</h3>
        <table>
          <tr v-for="item in platform"><td>{{ item }}</td></tr>
        </table>
      </div>

      <div class="col-lg-4">
        <h3>Bot</h3>
      </div>
      
      <div class="col-lg-4">
        <h3>Market</h3>
      </div>
      
    </div>
  </div>
</template>


<script>
export default {
  data () {
    return {
      platform: []
    }
  },
  mounted () {
    var self = this;
    this.ws = new WebSocket('ws://' + window.location.host + ':3028/messages');
    this.ws.addEventListener('message', function(e) {
        var msg = JSON.parse(e.data);
        console.log(e.data)
    });
  },
  methods: {
    getLogs (topic) {
      this.$http.get('http://localhost:3026/message/platform')
      .then(response => {
        this.platform = response.body
      }, response => {
        console.log('err', response)
      })
    }
  }
}
</script>
