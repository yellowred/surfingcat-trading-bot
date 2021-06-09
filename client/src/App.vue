<template>
<div id="app">

<nav class="navbar navbar-dark bg-dark">
  <a class="navbar-brand" href="/">SF Trading Bots</a>
  <a class="nav-link active" href="/">Home</a>
  <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNavAltMarkup" aria-controls="navbarNavAltMarkup" aria-expanded="false" aria-label="Toggle navigation">
    <span class="navbar-toggler-icon"></span>
  </button>
  <div class="collapse navbar-collapse" id="navbarNavAltMarkup">
    <div class="navbar-nav">
      <a class="nav-item nav-link active" href="#">Home <span class="sr-only">(current)</span></a>
      <a class="nav-item nav-link" href="#">Features</a>
      <a class="nav-item nav-link" href="#">Pricing</a>
      <a class="nav-item nav-link disabled" href="#">Disabled</a>
    </div>
  </div>
</nav>

  <div class="container">
    <router-view></router-view>
  </div>
</div>
</template>

<script>
import auth from './auth'

export default {

  data () {
    return {
    }
  },

  methods: {
    logout () {
      auth.logout(this)
    },
    isAuthenticated () {
      return auth.isAuthenticated()
    }
  }
}
</script>
