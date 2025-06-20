<template>
  <div>
    <div style="margin-bottom: 1em;">
      <button @click="fetchAll">Refresh</button>
      <button @click="allocateOrders">Allocate Orders</button>
    </div>

    <!-- Placement mode buttons -->
    <div style="margin-bottom: 1em;">
      <button @click="placementMode = 'warehouse'">Place Warehouse on Map</button>
      <button @click="placementMode = 'order'">Place Order on Map</button>
      <span v-if="placementMode">Click on the map to set {{ placementMode }} location</span>
    </div>

    <!-- Warehouse Creation Form -->
    <form @submit.prevent="createWarehouse" style="margin-bottom: 1em;">
      <h3>Add Warehouse</h3>
      <input v-model="newWarehouse.name" placeholder="Name" required />
      <input v-model.number="newWarehouse.lat" placeholder="Latitude" required type="number" step="any" />
      <input v-model.number="newWarehouse.lon" placeholder="Longitude" required type="number" step="any" />
      <button type="submit">Add Warehouse</button>
    </form>

    <!-- Agent Creation Form -->
    <form @submit.prevent="createAgent" style="margin-bottom: 1em;">
      <h3>Add Agent</h3>
      <input v-model="newAgent.name" placeholder="Name" required />
      <select v-model.number="newAgent.warehouse_id">
        <option value="">Unassigned</option>
        <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
      </select>
      <button type="submit">Add Agent</button>
    </form>

    <!-- Order Creation Form -->
    <form @submit.prevent="createOrder" style="margin-bottom: 1em;">
      <h3>Add Order</h3>
      <select v-model.number="newOrder.warehouse_id" required>
        <option disabled value="">Select Warehouse</option>
        <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
      </select>
      <input v-model="newOrder.delivery_address" placeholder="Delivery Address" required />
      <input v-model.number="newOrder.lat" placeholder="Latitude" required type="number" step="any" />
      <input v-model.number="newOrder.lon" placeholder="Longitude" required type="number" step="any" />
      <button type="submit">Add Order</button>
    </form>

    <div id="map" style="height: 500px; margin-bottom: 2em;"></div>

    <!-- Agents Table -->
    <div>
      <h3>Agents</h3>
      <table border="1" cellpadding="5" style="width:100%; text-align:left;">
        <thead>
          <tr>
            <th>Name</th>
            <th>Assigned Warehouse</th>
            <th>Payout</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="agent in agents" :key="agent.id">
            <td>{{ agent.name }}</td>
            <td>
              {{
                agent.warehouse_id
                  ? (warehouses.find(w => w.id === agent.warehouse_id)?.name || agent.warehouse_id)
                  : 'Unassigned'
              }}
            </td>
            <td>
              <span v-if="payouts[agent.id] !== undefined">₹{{ payouts[agent.id] }}</span>
              <span v-else>₹0</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import L from 'leaflet'
import axios from 'axios'

const warehouses = ref([])
const agents = ref([])
const orders = ref([])
const payouts = ref({}) // { agentId: payout }
let map

const placementMode = ref(null)

const newWarehouse = ref({ name: '', lat: '', lon: '' })
const newAgent = ref({ name: '', warehouse_id: '' })
const newOrder = ref({ warehouse_id: '', delivery_address: '', lat: '', lon: '' })

const initMap = async () => {
  await nextTick()
  if (!map) {
    map = L.map('map').setView([28.6139, 77.2090], 11)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors'
    }).addTo(map)

    map.on('click', (e) => {
      if (placementMode.value === 'warehouse') {
        newWarehouse.value.lat = e.latlng.lat
        newWarehouse.value.lon = e.latlng.lng
        placementMode.value = null
      } else if (placementMode.value === 'order') {
        newOrder.value.lat = e.latlng.lat
        newOrder.value.lon = e.latlng.lng
        placementMode.value = null
      }
    })
  }
}

const renderMap = () => {
  if (!map) return
  map.eachLayer(layer => {
    if (layer instanceof L.Marker || layer instanceof L.CircleMarker) map.removeLayer(layer)
  })

  // Warehouses
  warehouses.value.forEach(w => {
    L.marker([w.lat, w.lon], {
      icon: L.icon({
        iconUrl: 'https://cdn-icons-png.flaticon.com/512/684/684908.png',
        iconSize: [32, 32]
      })
    })
      .addTo(map)
      .bindPopup(`<b>Warehouse:</b> ${w.name}`)
  })

  // Orders
  orders.value.forEach(o => {
    L.circleMarker([o.lat, o.lon], { color: o.assigned ? 'green' : 'red', radius: 8 })
      .addTo(map)
      .bindPopup(`<b>Order:</b> ${o.delivery_address}<br>Assigned: ${o.assigned ? 'Yes' : 'No'}`)
  })
}

const fetchAll = async () => {
  try {
    warehouses.value = (await axios.get('/warehouses')).data
    agents.value = (await axios.get('/agents')).data
    orders.value = (await axios.get('/orders')).data
    // Reset payouts
    payouts.value = {}
  } catch (e) {
    console.error(e)
  }
  renderMap()
}

const allocateOrders = async () => {
  await axios.post('/allocate')
  await fetchAll()
  await fetchAllPayouts()
}

const fetchAllPayouts = async () => {
  const payoutMap = {}
  for (const agent of agents.value) {
    try {
      const payout = (await axios.get(`/agents/${agent.id}/payout`)).data
      payoutMap[agent.id] = payout.totalPay || 0
    } catch {
      payoutMap[agent.id] = 0
    }
  }
  payouts.value = payoutMap
}

const createWarehouse = async () => {
  await axios.post('/warehouses', {
    name: newWarehouse.value.name,
    lat: newWarehouse.value.lat,
    lon: newWarehouse.value.lon
  })
  newWarehouse.value = { name: '', lat: '', lon: '' }
  fetchAll()
}

const createAgent = async () => {
  await axios.post('/agents', {
    name: newAgent.value.name,
    warehouse_id: newAgent.value.warehouse_id || null
  })
  newAgent.value = { name: '', warehouse_id: '' }
  fetchAll()
}

const createOrder = async () => {
  await axios.post('/orders', {
    warehouse_id: newOrder.value.warehouse_id,
    delivery_address: newOrder.value.delivery_address,
    lat: newOrder.value.lat,
    lon: newOrder.value.lon
  })
  newOrder.value = { warehouse_id: '', delivery_address: '', lat: '', lon: '' }
  fetchAll()
}

onMounted(async () => {
  await initMap()
  fetchAll()
})
</script>