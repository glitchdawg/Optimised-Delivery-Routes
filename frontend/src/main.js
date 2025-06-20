import { createApp } from 'vue'
import App from './App.vue'
import axios from 'axios'

axios.defaults.baseURL = 'http://localhost:8080' // Change to your backend URL

createApp(App).mount('#app')