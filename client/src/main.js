// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import Home from '@/components/Home'
import Login from '@/components/Login'
import Logout from '@/components/Logout'
import Signup from '@/components/Signup'
import Balance from '@/components/Balance'
import Testbed from '@/components/Testbed'
import SecretQuote from '@/components/SecretQuote'
import UserInfo from '@/components/UserInfo'

import VueResource from 'vue-resource'
Vue.use(VueResource)

import VueRouter from 'vue-router'
Vue.use(VueRouter)

Vue.config.productionTip = true

import auth from './auth'

function requireAuth (to, from, next) {
  if (!auth.isAuthenticated()) {
    this.$router.replace('/login')
  } else {
    next()
  }
}

const router = new VueRouter({
  mode: 'history',
  // base: __dirname,
  routes: [
    {
      path: '/',
      component: Home
    },
    {
      path: '/home',
      name: 'home',
      component: Home
    },
    {
      path: '/login',
      name: 'login',
      component: Login
    },
    {
      path: '/logout',
      name: 'logout',
      component: Logout
    },
    {
      path: '/signup',
      name: 'signup',
      component: Signup
    },
    {
      path: '/testbed',
      name: 'testbed',
      component: Testbed
    },
    {
      path: '/balance',
      name: 'balance',
      component: Balance
    },
    {
      path: '/secretquote',
      name: 'secretquote',
      component: SecretQuote,
      beforeEnter: requireAuth
    },
    {
      path: '/userinfo',
      name: 'userinfo',
      component: UserInfo,
      beforeEnter: requireAuth
    }
  ]
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  template: '<App/>',
  components: { App }
})
