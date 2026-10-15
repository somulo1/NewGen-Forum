<template>
  <div class="roles-container">
    <div class="page-header">
      <div>
        <h1>Roles & Rights Management</h1>
        <p class="subtitle">Configure user roles and permissions</p>
      </div>
      <button class="btn btn-primary" @click="showNewRole = true">
        <Icon name="plus" size="18" />
        New Role
      </button>
    </div>

    <div class="roles-layout">
      <div class="roles-list">
        <div
          v-for="role in roles"
          :key="role.id"
          :class="['role-card', { active: selectedRole?.id === role.id }]"
          @click="selectedRole = role"
        >
          <div class="role-header">
            <h3>{{ role.name }}</h3>
            <button class="role-menu">
              <Icon name="more-vertical" size="16" />
            </button>
          </div>
          <p class="role-members">{{ role.members }} members</p>
          <div class="role-badges">
            <span v-for="perm in role.permissions.slice(0, 2)" :key="perm" class="perm-badge">{{ perm }}</span>
            <span v-if="role.permissions.length > 2" class="perm-badge">+{{ role.permissions.length - 2 }}</span>
          </div>
        </div>
      </div>

      <div v-if="selectedRole" class="role-details">
        <div class="details-header">
          <h2>{{ selectedRole.name }}</h2>
          <button class="edit-btn">
            <Icon name="edit" size="16" />
            Edit
          </button>
        </div>

        <div class="details-section">
          <h3>Members ({{ selectedRole.members }})</h3>
          <div class="members-list">
            <div v-for="member in selectedRole.memberList" :key="member.id" class="member-item">
              <div class="member-avatar">{{ member.initials }}</div>
              <div class="member-info">
                <p>{{ member.name }}</p>
                <span>{{ member.department }}</span>
              </div>
              <button class="remove-btn">Remove</button>
            </div>
          </div>
        </div>

        <div class="details-section">
          <h3>Permissions</h3>
          <div class="permissions-grid">
            <div v-for="perm in selectedRole.permissions" :key="perm" class="permission-item">
              <input type="checkbox" :checked="true" />
              <label>{{ perm }}</label>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="no-selection">
        <Icon name="lock" size="48" />
        <p>Select a role to view details</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const showNewRole = ref(false)
const selectedRole = ref(null)

const roles = [
  {
    id: 1,
    name: 'Administrator',
    members: 3,
    permissions: ['create_users', 'delete_users', 'manage_roles', 'view_logs', 'export_data', 'system_settings'],
    memberList: [
      { id: 1, name: 'Alice Johnson', initials: 'AJ', department: 'IT Operations' },
      { id: 2, name: 'Bob Smith', initials: 'BS', department: 'IT Operations' },
      { id: 3, name: 'Carol Davis', initials: 'CD', department: 'Security' },
    ],
  },
  {
    id: 2,
    name: 'Manager',
    members: 12,
    permissions: ['create_projects', 'manage_team', 'view_reports', 'approve_requests', 'edit_content'],
    memberList: [
      { id: 4, name: 'David Wilson', initials: 'DW', department: 'Sales' },
      { id: 5, name: 'Emma Taylor', initials: 'ET', department: 'Marketing' },
    ],
  },
  {
    id: 3,
    name: 'Team Lead',
    members: 8,
    permissions: ['create_tasks', 'assign_work', 'view_team_data', 'comment_on_issues'],
    memberList: [
      { id: 6, name: 'Frank Brown', initials: 'FB', department: 'Engineering' },
    ],
  },
  {
    id: 4,
    name: 'Employee',
    members: 85,
    permissions: ['view_documents', 'create_tickets', 'comment_on_tickets'],
    memberList: [],
  },
]
</script>

<style scoped>
.roles-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: var(--spacing-sm) 0 0 0;
}

.roles-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: var(--spacing-lg);
  min-height: 500px;
}

.roles-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.role-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  cursor: pointer;
  transition: all 0.2s ease;
}

.role-card:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.role-card.active {
  border-color: var(--color-lime);
  background: rgba(184, 243, 107, 0.05);
}

.role-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-md);
}

.role-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.role-menu {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 0;
}

.role-members {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.role-badges {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.perm-badge {
  font-size: 11px;
  padding: 2px 6px;
  background: var(--color-border);
  color: var(--color-text-secondary);
  border-radius: var(--radius-xs);
}

.role-details {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xl);
}

.details-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.details-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.edit-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  background: var(--color-border);
  border: none;
  color: var(--color-text);
  padding: var(--spacing-sm) var(--spacing-lg);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.edit-btn:hover {
  background: var(--color-lime);
  color: #000;
}

.details-section {
  margin-bottom: var(--spacing-xl);
}

.details-section h3 {
  margin: 0 0 var(--spacing-lg) 0;
  font-size: 14px;
  font-weight: 600;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.member-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.member-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 12px;
  flex-shrink: 0;
}

.member-info {
  flex: 1;
}

.member-info p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.member-info span {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
  display: block;
}

.remove-btn {
  background: rgba(255, 155, 138, 0.1);
  color: var(--color-coral);
  border: none;
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.remove-btn:hover {
  background: rgba(255, 155, 138, 0.2);
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.permission-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.permission-item input {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.permission-item label {
  margin: 0;
  font-size: 13px;
  color: var(--color-text);
  cursor: pointer;
  flex: 1;
}

.no-selection {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);
  gap: var(--spacing-lg);
}

.no-selection :deep(svg) {
  opacity: 0.5;
}

@media (max-width: 1024px) {
  .roles-layout {
    grid-template-columns: 1fr;
  }
}
</style>
