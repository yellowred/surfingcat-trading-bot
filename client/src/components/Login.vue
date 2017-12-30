<template>
<div class="row justify-content-center mt-4">
  <div class="justify-content-center col-sm-4 col-sm-offset-4">
    <h2>Log In</h2>
    <div class="alert alert-danger" v-if="error">
      <p>{{ error }}</p>
    </div>
    <div class="form-group">
      <input
        type="text"
        class="form-control"
        placeholder="Enter your username"
        v-model="credentials.login"
        v-on:keyup.enter="submit"
      >
    </div>
    <div class="form-group">
      <input
        type="password"
        class="form-control"
        placeholder="Enter your password"
        v-model="credentials.password"
        v-on:keyup.enter="submit"
      >
    </div>
    <button class="btn btn-primary" @click="submit">Login</button>
  </div>
</div>
</template>

<script>
import auth from '../auth'

export default {
  data () {
    return {
      credentials: {
        login: '',
        password: ''
      },
      error: ''
    }
  },
  methods: {
    submit () {
      var credentials = {
        login: this.credentials.login,
        password: this.credentials.password
      }
      auth.login(this, credentials, 'secretquote')
    }
  }
}
</script>
