<template>
  <div class="profile-container">
    <!-- Profile Header -->
    <div class="profile-header" style="background: linear-gradient(135deg, #4D5C3D 0%, #5C4D3D 100%)">
      <div class="header-content">
        <div class="profile-avatar" :style="{ background: profile.avatarColor }">{{ profile.initials }}</div>
        <div class="profile-info">
          <h1>{{ profile.name }}</h1>
          <p class="profile-title">{{ profile.title }} · {{ profile.department }}</p>
          <div class="status-select">
            <span class="status-dot"></span>
            <select class="status-dropdown">
              <option>Available</option>
              <option>Away</option>
              <option>Focusing</option>
              <option>In a meeting</option>
            </select>
          </div>
        </div>
      </div>

      <div class="profile-actions">
        <button class="btn-edit">Edit profile</button>
        <button class="btn-share">Share profile</button>
      </div>
    </div>

    <!-- Profile Completion -->
    <div class="completion-section">
      <p class="completion-text">Profile 60% complete</p>
      <div class="progress-bar">
        <div class="progress-fill" style="width: 60%"></div>
      </div>
      <p class="completion-hint">Add a profile photo so colleagues recognise you.</p>
    </div>

    <!-- Info Grid -->
    <div class="info-grid">
      <div class="info-card">
        <h3>Department</h3>
        <p class="info-value">{{ profile.department }}</p>
        <p class="info-detail">{{ profile.team }}</p>
      </div>
      <div class="info-card">
        <h3>Reports to</h3>
        <p class="info-value">{{ profile.reportsTo }}</p>
        <p class="info-detail">{{ profile.reportsToDept }}</p>
      </div>
      <div class="info-card">
        <h3>Location</h3>
        <p class="info-value">{{ profile.location }}</p>
        <p class="info-detail">{{ profile.timezone }}</p>
      </div>
      <div class="info-card">
        <h3>Member since</h3>
        <p class="info-value">{{ profile.joinDate }}</p>
        <p class="info-detail">Hub member</p>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-item">
        <div class="stat-number">{{ profile.ticketsRaised }}</div>
        <div class="stat-label">Tickets raised</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">{{ profile.ticketsHandled }}</div>
        <div class="stat-label">Tickets handled</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">{{ profile.wallPosts }}</div>
        <div class="stat-label">Wall posts</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">{{ profile.kudosReceived }}</div>
        <div class="stat-label">Kudos received</div>
      </div>
    </div>

    <!-- About Section -->
    <div class="about-section">
      <h2>About me</h2>
      <p>{{ profile.bio }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})

const profile = {
  name: 'Samwel Okoth Omulo',
  initials: 'S',
  avatarColor: '#F49AC1',
  title: 'Cybersecurity & Technology Risk Officer',
  department: 'Risk & Compliance',
  team: 'Cybersecurity team',
  reportsTo: 'Head of Risk Management',
  reportsToDept: 'Risk & Compliance',
  location: 'Kenya',
  timezone: 'GMT+3 · 23:40 now',
  joinDate: '[JOIN DATE]',
  ticketsRaised: 1,
  ticketsHandled: 0,
  wallPosts: 0,
  kudosReceived: 14,
  bio: 'I look after cybersecurity and technology risk across our systems, from access reviews and audits to making sure new tools go live safely. Happy to help with anything security related. Just tag me in a ticket or ping me in Risk & Compliance.',
}
</script>

<style scoped>
.profile-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.profile-header {
  border-radius: 12px;
  padding: 32px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 20px;
  flex: 1;
}

.profile-avatar {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
  position: relative;
}

.profile-avatar::after {
  content: '';
  position: absolute;
  bottom: 0;
  right: 0;
  width: 32px;
  height: 32px;
  background: #B8F36B;
  border-radius: 50%;
  border: 3px solid var(--color-bg);
}

.profile-info h1 {
  margin: 0 0 4px 0;
  font-size: 28px;
  font-weight: 600;
  color: white;
}

.profile-title {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.status-select {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 6px 12px;
  width: fit-content;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #B8F36B;
  border-radius: 50%;
}

.status-dropdown {
  background: transparent;
  border: none;
  color: white;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  padding: 0;
}

.status-dropdown:focus {
  outline: none;
}

.profile-actions {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.btn-edit {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 10px 20px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn-share {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 10px 20px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-share:hover {
  border-color: white;
}

.completion-section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.completion-text {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.progress-bar {
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 12px;
}

.progress-fill {
  height: 100%;
  background: #FFB86B;
  border-radius: 3px;
}

.completion-hint {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.info-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
}

.info-card h3 {
  margin: 0 0 8px 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
}

.info-value {
  margin: 0 0 4px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.info-detail {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-item {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  color: #FFB86B;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.about-section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.about-section h2 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.about-section p {
  margin: 0;
  font-size: 13px;
  color: var(--color-text);
  line-height: 1.6;
}

@media (max-width: 1024px) {
  .profile-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-content {
    width: 100%;
  }

  .profile-actions {
    width: 100%;
  }

  .btn-edit,
  .btn-share {
    flex: 1;
  }

  .info-grid,
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .info-grid,
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
