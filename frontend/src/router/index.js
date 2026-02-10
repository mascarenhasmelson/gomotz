
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('../views/Home.vue')
    },
    {
      path: '/portforward',
      component: () => import('../views/Portforward.vue')
    },
    {
      path: '/tools',
      component: () => import('../views/NetworkTools.vue'),
      children: [
        {
          path: 'portscan',
          component: () => import('../views/tools/Portscan.vue')
        },
        {
          path: 'tcp-check',
          component: () => import('../views/tools/TcpCheck.vue')
        },
        {
          path: 'dns-lookup',
          component: () => import('../views/tools/DnsLookup.vue')
        },
        {
          path: 'traceroute',
          component: () => import('../views/tools/Traceroute.vue')
        },
        {
          path: 'ping',
          component: () => import('../views/tools/Ping.vue')
        },
        {
          path: 'HttpsCheck',
          component: () => import('../views/tools/HttpsCheck.vue')
        },
        {
          path: 'speedtest',
          component: () => import('../views/tools/Speedtest.vue')
        },
      ]
    },
    {
      path: '/monitor',
      component: () => import('../views/Monitor.vue'),
      children: [
        {
          path: 'ARP',
          component: () => import('../views/monitor/ARP.vue')
        },
    
        {
          path: 'HTTPsMonitoring',
          component: () => import('../views/monitor/HTTPsMonitoring.vue')
        },
      
        {
          path: 'SMTPMonitoring',
          component: () => import('../views/monitor/SMTPMonitoring.vue')
        },
         {
          path: 'SNMPMonitor',
          component: () => import('../views/monitor/SNMPMonitor.vue')
        },
             {
          path: 'PingMonitor',
          component: () => import('../views/monitor/PingMonitor.vue')
        },
          {
          path: 'HTTPsKeyword',
          component: () => import('../views/monitor/HTTPsKeyword.vue')
        },
         {
          path: 'HTTPsJSON',
          component: () => import('../views/monitor/HTTPsJSON.vue')
        },
           {
          path: 'gRPCJSON',
          component: () => import('../views/monitor/gRPCJSON.vue')
        },
        {
          path: 'TCPMonitoring',
          component: () => import('../views/monitor/TCPMonitoring.vue')
        },
       
         {
          path: 'LANWakeup',
          component: () => import('../views/monitor/LanWakeup.vue')
        },
          
        {
          path: 'domain-expiry',
          component: () => import('../views/monitor/DomainExpiry.vue')
        },
      ]
    },
     {
      path: '/remote',
      component: () => import('../views/RemoteAccess.vue'),
      children: [
        {
          path: 'ssh',
          component: () => import('../views/remote/ssh.vue')
        },
        {
          path: 'telnet',
          component: () => import('../views/remote/telnet.vue')
        },
    
      ]
    }



  ]
})

export default router