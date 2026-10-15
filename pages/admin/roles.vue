<template>
  <div class="roles-container">
    <div class="page-header">
      <div>
        <h1>Roles & rights</h1>
        <p class="subtitle">Who can do what in Hub</p>
      </div>
    </div>

    <div class="roles-layout">
      <div class="roles-sidebar">
        <h3 class="sidebar-title">Roles</h3>
        <div v-for="role in roles" :key="role.id" :class="['role-item', { active: selectedRoleId === role.id }]" @click="selectedRoleId = role.id">
          {{ role.name }}
        </div>
      </div>

      <div class="role-details">
        <div v-if="selectedRole" class="details-content">
          <h2>{{ selectedRole.name }}</h2>

          <div class="members-section">
            <h3>Members</h3>
            <div class="members-list">
              <div v-for="member in selectedRole.members" :key="member.id" class="member-item">
                <div class="member-avatar" :style="{ background: member.color }">{{ member.initials }}</div>
                <span class="member-name">{{ member.name }}</span>
              </div>
            </div>
          </div>

          <div class="permissions-section">
            <h3>Permissions</h3>
            <div class="permissions-list">
              <div v-for="permission in selectedRole.permissions" :key="permission.id" class="permission-item">
                <input type="checkbox" :id="`perm-${permission.id}`" :checked="permission.enabled">
                <label :for="`perm-${permission.id}`">{{ permission.name }}</label>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'default',
})

const selectedRoleId = ref(1)

const roles = [
  {
    id: 1,
    name: 'Employee',
    members: [
      { id: 1, name: 'John Doe', initials: 'JD', color: '#7CC4FF' },
      { id: 2, name: 'Sarah Martin', initials: 'SM', color: '#F49AC1' },
      { id: 3, name: 'Marcus Davis', initials: 'MD', color: '#FFB86B' },
    ],
    permissions: [
      { id: 1, name: 'View announcements', enabled: true },
      { id: 2, name: 'View people directory', enabled: true },
      { id: 3, name: 'Post to community wall', enabled: true },
      { id: 4, name: 'Create tickets', enabled: true },
      { id: 5, name: 'View calendar', enabled: true },
      { id: 6, name: 'Approve requests', enabled: false },
      { id: 7, name: 'Manage security', enabled: false },
      { id: 8, name: 'Delete content', enabled: false },
    ],
  },
  {
    id: 2,
    name: 'Manager',
    members: [
      { id: 4, name: 'Grace Njeri', initials: 'GN', color: '#FFB86B' },
      { id: 5, name: 'David Otieno', initials: 'DO', color: '#F49AC1' },
    ],
    permissions: [
      { id: 1, name: 'View announcements', enabled: true },
      { id: 2, name: 'View people directory', enabled: true },
      { id: 3, name: 'Post to community wall', enabled: true },
      { id: 4, name: 'Create tickets', enabled: true },
      { id: 5, name: 'View calendar', enabled: true },
      { id: 6, name: 'Approve requests', enabled: true },
      { id: 7, name: 'Manage security', enabled: false },
      { id: 8, name: 'Delete content', enabled: false },
    ],
  },
  {
    id: 3,
    name: 'Administrator',
    members: [
      { id: 6, name: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8' },
    ],
    permissions: [
      { id: 1, name: 'View announcements', enabled: true },
      { id: 2, name: 'View people directory', enabled: true },
      { id: 3, name: 'Post to community wall', enabled: true },
      { id: 4, name: 'Create tickets', enabled: true },
      { id: 5, name: 'View calendar', enabled: true },
      { id: 6, name: 'Approve requests', enabled: true },
      { id: 7, name: 'Manage security', enabled: true },
      { id: 8, name: 'Delete content', enabled: true },
    ],
  },
]

const selectedRole = computed(() => roles.find(r => r.id === selectedRoleId.value))
</script>

<style scoped>
.roles-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: 8px 0 0 0;
}

.roles-layout {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 24px;
}

.roles-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.sidebar-title {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text-secondary);
  text-transform: uppercase;
}

.role-item {
  padding: 12px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.role-item:hover {
  border-color: #B8F36B;
  color: var(--color-text);
}

.role-item.active {
  background: var(--color-surface);
  border-color: #B8F36B;
  color: #B8F36B;
  font-weight: 600;
}

.role-details {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 24px;
}

.details-content h2 {
  margin: 0 0 24px 0;
  font-size: 20px;
  font-weight: 600;
}

.members-section,
.permissions-section {
  margin-bottom: 32px;
}

.members-section h3,
.permissions-section h3 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.member-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: white;
}

.member-name {
  font-size: 13px;
  color: var(--color-text);
}

.permissions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.permission-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.permission-item input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.permission-item label {
  font-size: 13px;
  color: var(--color-text);
  cursor: pointer;
}

@media (max-width: 768px) {
  .roles-layout {
    grid-template-columns: 1fr;
  }
}
</style>
