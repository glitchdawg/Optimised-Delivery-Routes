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
      <div v-if="newWarehouse.lat && newWarehouse.lon" style="margin-top:0.5em; color: #333;">
        <b>Selected Location:</b> {{ newWarehouse.lat }}, {{ newWarehouse.lon }}
      </div>
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
      <div v-if="newOrder.lat && newOrder.lon" style="margin-top:0.5em; color: #333;">
        <b>Selected Location:</b> {{ newOrder.lat }}, {{ newOrder.lon }}
      </div>
    </form>

    <div id="map" style="height: 500px; margin-bottom: 2em;"></div>

    <!-- Metrics Section -->
    <div style="margin-bottom: 2em;">
      <h3>Metrics</h3>
      <ul>
        <li>Total Warehouses: {{ warehouses.length }}</li>
        <li>Total Agents: {{ agents.length }}</li>
        <li>Total Orders: {{ orders.length }}</li>
        <li>Assigned Orders: {{ assignedOrdersCount }}</li>
        <li>Unassigned Orders: {{ unassignedOrdersCount }}</li>
        <li>Average Orders per Agent: {{ avgOrdersPerAgent }}</li>
        <li>Average Distance per Agent: {{ avgDistancePerAgent.toFixed(2) }} km</li>
        <li>Total Cost Today: ₹{{ totalCost }}</li>
      </ul>
    </div>

    <!-- Agents Table -->
    <div>
      <h3>Agents</h3>
      <table border="1" cellpadding="5" style="width:100%; text-align:left;">
        <thead>
          <tr>
            <th>Name</th>
            <th>Assigned Warehouse</th>
            <th>Orders Assigned</th>
            <th>Total Distance (km)</th>
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
            <td>{{ agentOrders[agent.id]?.length || 0 }}</td>
            <td>{{ agentDistances[agent.id]?.toFixed(2) || 0 }}</td>
            <td>
              <span v-if="payouts[agent.id] !== undefined">₹{{ payouts[agent.id] }}</span>
              <span v-else>₹0</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Orders Table -->
    <div style="margin-top:2em;">
      <h3>Orders</h3>
      <table border="1" cellpadding="5" style="width:100%; text-align:left;">
        <thead>
          <tr>
            <th>ID</th>
            <th>Warehouse</th>
            <th>Address</th>
            <th>Assigned</th>
            <th>Agent</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in orders" :key="o.id">
            <td>{{ o.id }}</td>
            <td>{{ warehouses.find(w => w.id === o.warehouse_id)?.name || o.warehouse_id }}</td>
            <td>{{ o.delivery_address }}</td>
            <td>{{ o.assigned ? 'Yes' : 'No' }}</td>
            <td>{{ agentNameById(o.agent_id) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, computed, watch } from 'vue'
import L from 'leaflet'
import axios from 'axios'

const warehouses = ref([])
const agents = ref([])
const orders = ref([])
const payouts = ref({})
const agentOrders = ref({})
const agentDistances = ref({})
let map

const placementMode = ref(null)

const newWarehouse = ref({ name: '', lat: '', lon: '' })
const newAgent = ref({ name: '', warehouse_id: '' })
const newOrder = ref({ warehouse_id: '', delivery_address: '', lat: '', lon: '' })

const assignedOrdersCount = computed(() => orders.value.filter(o => o.assigned).length)
const unassignedOrdersCount = computed(() => orders.value.filter(o => !o.assigned).length)
const avgOrdersPerAgent = computed(() => agents.value.length ? (assignedOrdersCount.value / agents.value.length).toFixed(2) : 0)
const avgDistancePerAgent = computed(() => {
  const total = Object.values(agentDistances.value).reduce((a, b) => a + b, 0)
  return agents.value.length ? total / agents.value.length : 0
})
const totalCost = computed(() => Object.values(payouts.value).reduce((a, b) => a + b, 0))

const agentNameById = (id) => {
  if (!id) return ''
  const agent = agents.value.find(a => a.id === id)
  return agent ? agent.name : id
}

const renderMap = () => {
  if (!map) return
  // Remove only markers and circle markers, not the tile layer
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
      .bindPopup(`<b>Warehouse:</b> ${w.name}<br>Lat: ${w.lat}<br>Lon: ${w.lon}`)
  })

  // Orders
  orders.value.forEach(o => {
    L.circleMarker([o.lat, o.lon], { color: o.assigned ? 'green' : 'red', radius: 8 })
      .addTo(map)
      .bindPopup(`<b>Order:</b> ${o.delivery_address}<br>Assigned: ${o.assigned ? 'Yes' : 'No'}<br>Lat: ${o.lat}<br>Lon: ${o.lon}`)
  })
}

const fetchAll = async () => {
  try {
    const w = (await axios.get('/warehouses')).data
    const a = (await axios.get('/agents')).data
    const o = (await axios.get('/orders')).data
    warehouses.value = Array.isArray(w) ? w : []
    agents.value = Array.isArray(a) ? a : []
    orders.value = Array.isArray(o) ? o : []
    await fetchAgentOrdersAndDistances()
    await fetchAllPayouts()
    await nextTick()
    renderMap()
  } catch (e) {
    console.error(e)
  }
}

const allocateOrders = async () => {
  await axios.post('/allocate')
  await fetchAll()
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

const fetchAgentOrdersAndDistances = async () => {
  const ordersMap = {}
  const distancesMap = {}
  for (const agent of agents.value) {
    try {
      const agentOrdersRes = await axios.get(`/agents/${agent.id}/orders`)
      ordersMap[agent.id] = agentOrdersRes.data || []
      let totalDist = 0
      for (const o of agentOrdersRes.data) {
        if (o.distance_km) {
          totalDist += o.distance_km
        }
      }
      distancesMap[agent.id] = totalDist
    } catch {
      ordersMap[agent.id] = []
      distancesMap[agent.id] = 0
    }
  }
  agentOrders.value = ordersMap
  agentDistances.value = distancesMap
}

const createWarehouse = async () => {
  await axios.post('/warehouses', {
    name: newWarehouse.value.name,
    lat: newWarehouse.value.lat,
    lon: newWarehouse.value.lon
  })
  newWarehouse.value = { name: '', lat: '', lon: '' }
  await fetchAll()
}

const createAgent = async () => {
  await axios.post('/agents', {
    name: newAgent.value.name,
    warehouse_id: newAgent.value.warehouse_id || null
  })
  newAgent.value = { name: '', warehouse_id: '' }
  await fetchAll()
}

const createOrder = async () => {
  await axios.post('/orders', {
    warehouse_id: newOrder.value.warehouse_id,
    delivery_address: newOrder.value.delivery_address,
    lat: newOrder.value.lat,
    lon: newOrder.value.lon
  })
  newOrder.value = { warehouse_id: '', delivery_address: '', lat: '', lon: '' }
  await fetchAll()
}

const initMap = async () => {
  await nextTick()
  if (!map) {
    map = L.map('map').setView([28.6139, 77.2090], 11)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '© OpenStreetMap contributors'
    }).addTo(map)

    map.on('click', (e) => {
      if (placementMode.value === 'warehouse') {
        newWarehouse.value.lat = Number(e.latlng.lat.toFixed(6))
        newWarehouse.value.lon = Number(e.latlng.lng.toFixed(6))
        placementMode.value = null
      } else if (placementMode.value === 'order') {
        newOrder.value.lat = Number(e.latlng.lat.toFixed(6))
        newOrder.value.lon = Number(e.latlng.lng.toFixed(6))
        placementMode.value = null
      }
    })
  }
  await nextTick()
  renderMap()
}

// Re-render map if warehouses or orders change and map is ready
watch([warehouses, orders], () => {
  if (map) nextTick().then(renderMap)
})

onMounted(async () => {
  await initMap()
  await fetchAll()
})
</script>