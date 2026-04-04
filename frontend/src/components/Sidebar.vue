<template>
	<aside :class="['sidebar', { 'is-expanded': is_expanded, 'is-collapsed': !is_expanded }]">
		<div class="sidebar-header">
			<div class="logo-wrapper">
				<img :src="logoURL" alt="Logo" class="logo-img" />
				<transition name="fade">
					<h1 class="logo-text" v-if="is_expanded">Gomotz</h1>
				</transition>
			</div>
			
			<button class="menu-toggle" @click="ToggleMenu" :title="is_expanded ? 'Collapse sidebar' : 'Expand sidebar'">
				<svg class="toggle-icon" viewBox="0 0 24 24" width="24" height="24">
					<path :d="is_expanded ? 'M15.41 16.59L10.83 12l4.58-4.59L14 6l-6 6 6 6 1.41-1.41z' : 'M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z'" 
						fill="currentColor"/>
				</svg>
			</button>
		</div>

		<div class="sidebar-content">
			<nav class="main-nav">
				<div class="nav-section">
					<div class="section-label" v-if="is_expanded">Navigation</div>
					
					<router-link to="/" class="nav-item" active-class="active" exact>
						<div class="nav-icon">
							<svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
								<path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z" 
									fill="currentColor"/>
							</svg>
						</div>
						<transition name="slide">
							<span class="nav-text" v-if="is_expanded">Dashboard</span>
						</transition>
						<div class="active-indicator"></div>
					</router-link>
					
					<router-link to="/portforward" class="nav-item" active-class="active">
						<div class="nav-icon">
							<svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
								<path d="M6.99 11L3 15l3.99 4v-3H14v-2H6.99v-3zM21 9l-3.99-4v3H10v2h7.01v3L21 9z" 
									fill="currentColor"/>
							</svg>
						</div>
						<transition name="slide">
							<span class="nav-text" v-if="is_expanded">Port Forwarding</span>
						</transition>
						<div class="active-indicator"></div>
					</router-link>
					
					<router-link to="/tools" class="nav-item" active-class="active">
						<div class="nav-icon">
							<svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
								<path d="M22.7 19l-9.1-9.1c.9-2.3.4-5-1.5-6.9-2-2-5-2.4-7.4-1.3L9 6 6 9 1.6 4.7C.4 7.1.9 10.1 2.9 12.1c1.9 1.9 4.6 2.4 6.9 1.5l9.1 9.1c.4.4 1 .4 1.4 0l2.3-2.3c.5-.4.5-1.1.1-1.4z" 
									fill="currentColor"/>
							</svg>
						</div>
						<transition name="slide">
							<span class="nav-text" v-if="is_expanded">Network Tools</span>
						</transition>
						<div class="active-indicator"></div>
					</router-link>
					
<router-link to="/monitor" class="nav-item" active-class="active">
  <div class="nav-icon">
    <svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
      <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z" 
        fill="currentColor"/>
      <path d="M7 12l2-3 2 2 2-4 2 5 2-2 2 3" 
        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M7 16h10" 
        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      <circle cx="7" cy="10" r="1" fill="currentColor" opacity="0.8"/>
      <circle cx="11" cy="9" r="1" fill="currentColor" opacity="0.8"/>
      <circle cx="15" cy="11" r="1" fill="currentColor" opacity="0.8"/>
      <circle cx="19" cy="14" r="1" fill="currentColor" opacity="0.8"/>
    </svg>
  </div>
  <transition name="slide">
    <span class="nav-text" v-if="is_expanded">Monitoring</span>
  </transition>
  <div class="active-indicator"></div>
</router-link>

<router-link to="/remote" class="nav-item" active-class="active">
  <div class="nav-icon">
    <svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
      <path d="M4 6h18V4H4c-1.1 0-2 .9-2 2v11H0v3h14v-3H4V6zm19 2h-6c-.55 0-1 .45-1 1v10c0 .55.45 1 1 1h6c.55 0 1-.45 1-1V9c0-.55-.45-1-1-1zm-1 9h-4v-7h4v7z" 
        fill="currentColor"/>
      <circle cx="17.5" cy="11.5" r="1.5" fill="currentColor"/>
    </svg>
  </div>
  <transition name="slide">
    <span class="nav-text" v-if="is_expanded">Remote Access</span>
  </transition>
  <div class="active-indicator"></div>
</router-link>

    <router-link to="/block" class="nav-item block-item" active-class="active">
      <div class="nav-icon">
        <!-- Shield with Block Symbol - Main Icon -->
        <svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
          <!-- Shield Background -->
          <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z" 
                fill="currentColor" opacity="0.9"/>
          <!-- Block Symbol (X) -->
          <path d="M15.5 8.5L8.5 15.5M8.5 8.5L15.5 15.5" 
                stroke="currentColor" 
                stroke-width="2" 
                stroke-linecap="round"
                fill="none"/>
          <!-- Lock Icon (Optional detail) -->
          <circle cx="12" cy="13" r="2" fill="white" stroke="currentColor" stroke-width="1"/>
          <path d="M10 11V9c0-1.1.9-2 2-2s2 .9 2 2v2" 
                stroke="white" 
                stroke-width="1.5" 
                fill="none"/>
        </svg>
      </div>
      <transition name="slide">
        <span class="nav-text" v-if="is_expanded">
          <span class="block-text">Internet Block</span>
        </span>
      </transition>
      <div class="active-indicator"></div>
    </router-link>
				</div>
			</nav>

			<div class="spacer"></div>

			<nav class="secondary-nav">
				<div class="nav-section">
					 <div class="section-label" v-if="is_expanded">System</div> 
					
					 <router-link to="/settings" class="nav-item" active-class="active">
						<div class="nav-icon">
							<svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
								<path d="M19.43 12.98c.04-.32.07-.64.07-.98s-.03-.66-.07-.98l2.11-1.65c.19-.15.24-.42.12-.64l-2-3.46c-.12-.22-.39-.3-.61-.22l-2.49 1c-.52-.4-1.08-.73-1.69-.98l-.38-2.65C14.46 2.18 14.25 2 14 2h-4c-.25 0-.46.18-.49.42l-.38 2.65c-.61.25-1.17.59-1.69.98l-2.49-1c-.22-.09-.49 0-.61.22l-2 3.46c-.13.22-.07.49.12.64l2.11 1.65c-.04.32-.07.65-.07.98s.03.66.07.98l-2.11 1.65c-.19.15-.24.42-.12.64l2 3.46c.12.22.39.3.61.22l2.49-1c.52.4 1.08.73 1.69.98l.38 2.65c.03.24.24.42.49.42h4c.25 0 .46-.18.49-.42l.38-2.65c.61-.25 1.17-.59 1.69-.98l2.49 1c.22.08.49 0 .61-.22l2-3.46c.12-.22.07-.49-.12-.64l-2.11-1.65zM12 15.5c-1.93 0-3.5-1.57-3.5-3.5s1.57-3.5 3.5-3.5 3.5 1.57 3.5 3.5-1.57 3.5-3.5 3.5z" 
									fill="currentColor"/>
							</svg>
						</div>
						<transition name="slide">
							<span class="nav-text" v-if="is_expanded">Settings</span>
						</transition>
						<div class="active-indicator"></div>
					</router-link>
					
					<!-- <a href="#" class="nav-item" @click.prevent="logout">
						<div class="nav-icon">
							<svg class="icon-svg" viewBox="0 0 24 24" width="24" height="24">
								<path d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z" 
									fill="currentColor"/>
							</svg>
						</div>
						<transition name="slide">
							<span class="nav-text" v-if="is_expanded">Logout</span>
						</transition>
					</a>  -->
				</div>
			</nav>
		</div>

		<div class="sidebar-footer" v-if="is_expanded">
			<div class="status-indicator">
				<svg class="status-icon" viewBox="0 0 24 24" width="16" height="16">
					<circle cx="12" cy="12" r="4" fill="#10b981"/>
					<circle cx="12" cy="12" r="11" fill="none" stroke="#10b981" stroke-width="2" 
						stroke-opacity="0.3"/>
				</svg>
				<span class="status-text">System Online</span>
			</div>
			<div class="version-info">v0.5.5</div>
		</div>
		
		<!-- Mobile Overlay -->
		<div class="mobile-overlay" v-if="isMobile && is_expanded" @click="ToggleMenu"></div>
	</aside>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import logoURL from '../assets/logo.png'

const router = useRouter()
const is_expanded = ref(localStorage.getItem("sidebar_expanded") === "true")
const isMobile = ref(false)

const ToggleMenu = () => {
	is_expanded.value = !is_expanded.value
	localStorage.setItem("sidebar_expanded", is_expanded.value)
}

const checkMobile = () => {
	isMobile.value = window.innerWidth < 1024
	if (isMobile.value) {
		is_expanded.value = false
	}
}

const logout = () => {
	// Add your logout logic here
	console.log('Logout clicked')
	router.push('/login')
}

onMounted(() => {
	checkMobile()
	window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
	window.removeEventListener('resize', checkMobile)
})
</script>

<style lang="scss" scoped>
.sidebar {
	--sidebar-width-expanded: 280px;
	--sidebar-width-collapsed: 80px;
	--sidebar-bg: #0f172a;
	--sidebar-bg-hover: #1e293b;
	--sidebar-text: #cbd5e1;
	--sidebar-text-hover: #ffffff;
	--sidebar-active: #3b82f6;
	--sidebar-border: #334155;
	--sidebar-transition: 0.3s cubic-bezier(0.4, 0, 0.2, 1);
	
	position: fixed;
	top: 0;
	left: 0;
	height: 100vh;
	background: var(--sidebar-bg);
	color: var(--sidebar-text);
	z-index: 1000;
	display: flex;
	flex-direction: column;
	transition: width var(--sidebar-transition), transform var(--sidebar-transition);
	width: var(--sidebar-width-expanded);
	border-right: 1px solid var(--sidebar-border);
	overflow: hidden;
	box-shadow: 4px 0 20px rgba(0, 0, 0, 0.2);

	&.is-collapsed {
		width: var(--sidebar-width-collapsed);
		
		.sidebar-header .logo-text {
			opacity: 0;
			width: 0;
		}
		
		.section-label,
		.nav-text,
		.status-text,
		.version-info {
			opacity: 0;
			width: 0;
			height: 0;
			margin: 0;
			padding: 0;
			overflow: hidden;
		}
		
		.sidebar-footer {
			padding: 1rem 0.5rem;
		}
	}

	&.is-expanded {
		width: var(--sidebar-width-expanded);
	}
}

.sidebar-header {
	padding: 1.5rem 1.25rem;
	border-bottom: 1px solid var(--sidebar-border);
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;

	.logo-wrapper {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex: 1;
		min-width: 0;
	}

	.logo-img {
		width: 40px;
		height: 40px;
		border-radius: 8px;
		object-fit: contain;
		flex-shrink: 0;
	}

	.logo-text {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--sidebar-text-hover);
		margin: 0;
		white-space: nowrap;
		overflow: hidden;
		transition: all var(--sidebar-transition);
		background: linear-gradient(135deg, #60a5fa, #8b5cf6);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.menu-toggle {
		background: var(--sidebar-bg-hover);
		border: 1px solid var(--sidebar-border);
		border-radius: 8px;
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.2s;
		flex-shrink: 0;
		color: var(--sidebar-text);
		padding: 0;

		&:hover {
			background: var(--sidebar-active);
			color: white;
			transform: scale(1.05);
			box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
		}

		&:focus {
			outline: 2px solid var(--sidebar-active);
			outline-offset: 2px;
		}

		.toggle-icon {
			width: 1.25rem;
			height: 1.25rem;
			transition: transform var(--sidebar-transition);
		}
	}
}

.sidebar-content {
	flex: 1;
	padding: 1.5rem 0;
	display: flex;
	flex-direction: column;
	overflow-y: auto;
	overflow-x: hidden;
	scrollbar-width: thin;
	scrollbar-color: var(--sidebar-border) transparent;

	&::-webkit-scrollbar {
		width: 6px;
	}

	&::-webkit-scrollbar-track {
		background: transparent;
	}

	&::-webkit-scrollbar-thumb {
		background: var(--sidebar-border);
		border-radius: 3px;
	}

	&::-webkit-scrollbar-thumb:hover {
		background: var(--sidebar-active);
	}
}

.spacer {
	flex: 1;
	min-height: 1rem;
}

.nav-section {
	margin-bottom: 2rem;

	.section-label {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #94a3b8;
		padding: 0.5rem 1.5rem 0.5rem;
		margin-bottom: 0.5rem;
		font-weight: 600;
		transition: all var(--sidebar-transition);
		white-space: nowrap;
		overflow: hidden;
	}
}

.nav-item {
	display: flex;
	align-items: center;
	gap: 1rem;
	padding: 0.875rem 1.5rem;
	margin: 0.25rem 0.75rem;
	text-decoration: none;
	color: var(--sidebar-text);
	border-radius: 8px;
	transition: all 0.2s;
	position: relative;
	overflow: hidden;

	&:hover {
		background: var(--sidebar-bg-hover);
		color: var(--sidebar-text-hover);
		transform: translateX(4px);

		.nav-icon {
			transform: scale(1.1);
		}
	}

	&.active {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(139, 92, 246, 0.15));
		color: var(--sidebar-active);
		border-left: 3px solid var(--sidebar-active);

		.nav-icon {
			color: var(--sidebar-active);
		}

		&::before {
			content: '';
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;
			background: linear-gradient(90deg, transparent, rgba(59, 130, 246, 0.1), transparent);
			animation: shimmer 2s infinite;
		}

		.active-indicator {
			opacity: 1;
			transform: scale(1);
		}
	}

	.nav-icon {
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: transform 0.2s;

		.icon-svg {
			width: 100%;
			height: 100%;
			color: currentColor;
		}
	}

	.nav-text {
		font-size: 0.95rem;
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		transition: all var(--sidebar-transition);
		flex: 1;
		min-width: 0;
	}

	.active-indicator {
		position: absolute;
		right: 12px;
		width: 8px;
		height: 8px;
		background: var(--sidebar-active);
		border-radius: 50%;
		opacity: 0;
		transform: scale(0);
		transition: all 0.2s;
		box-shadow: 0 0 10px var(--sidebar-active);
	}
}

.sidebar-footer {
	padding: 1rem 1.5rem;
	border-top: 1px solid var(--sidebar-border);
	background: rgba(15, 23, 42, 0.8);
	backdrop-filter: blur(10px);

	.status-indicator {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.75rem;

		.status-icon {
			flex-shrink: 0;
			animation: pulse 2s infinite;
			
			circle:nth-child(2) {
				animation: rotate 4s linear infinite;
				transform-origin: center;
			}
		}

		.status-text {
			font-size: 0.85rem;
			color: #94a3b8;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			transition: all var(--sidebar-transition);
		}
	}

	.version-info {
		font-size: 0.75rem;
		color: #64748b;
		text-align: center;
		transition: all var(--sidebar-transition);
	}
}

/* Animations */
@keyframes shimmer {
	0% {
		transform: translateX(-100%);
	}
	100% {
		transform: translateX(100%);
	}
}

@keyframes pulse {
	0%, 100% {
		opacity: 1;
	}
	50% {
		opacity: 0.7;
	}
}

@keyframes rotate {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(360deg);
	}
}

/* Transition animations */
.fade-enter-active,
.fade-leave-active {
	transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
	opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
	transition: all var(--sidebar-transition);
}

.slide-enter-from,
.slide-leave-to {
	opacity: 0;
	transform: translateX(-10px);
}

/* Mobile Overlay */
.mobile-overlay {
	position: fixed;
	top: 0;
	left: var(--sidebar-width-expanded);
	right: 0;
	bottom: 0;
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(4px);
	z-index: 999;
	animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
	from {
		opacity: 0;
	}
	to {
		opacity: 1;
	}
}

/* Responsive Design */
@media (max-width: 1024px) {
	.sidebar {
		transform: translateX(-100%);
		width: var(--sidebar-width-expanded);
		box-shadow: 4px 0 40px rgba(0, 0, 0, 0.3);

		&.is-expanded {
			transform: translateX(0);
		}

		&.is-collapsed {
			transform: translateX(-100%);
			width: var(--sidebar-width-expanded);
		}
	}
}

@media (max-width: 768px) {
	.sidebar {
		--sidebar-width-expanded: 260px;
	}
}

@media (min-width: 1025px) {
	.sidebar {
		position: sticky;
		top: 0;
		height: 100vh;
	}
}

/* Dark mode adjustments */
@media (prefers-color-scheme: dark) {
	.sidebar {
		--sidebar-bg: #0a0a0a;
		--sidebar-bg-hover: #1a1a1a;
		--sidebar-border: #2a2a2a;
	}
}
</style>