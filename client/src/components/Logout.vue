<template>
<div class="row justify-content-center mt-4">
  <div class="justify-content-center col-sm-4 col-sm-offset-4">
    <h2>Logout</h2>
      <button class="btn btn-primary" @click="submit">Logout</button>
  </div>
</div>
</template>

<script>
import auth from '../auth'

export default {

  methods: {
    submit () {
      auth.logout(this)
    }
  }
}
</script>
