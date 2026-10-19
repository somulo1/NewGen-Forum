import { useAccounts } from '~/composables/useAccounts';

export const peopleAdminActions = {
  updatePeopleAdmin(id, patch) {
    const person = this.dir.find((p) => p.id === id);
    if (!person) {
      this.notify('That person is no longer in the directory.', '#FF9B8A');
      this.setState({ userActionOpen: null, userResetTemp: '' });
      return;
    }
    const current = this.state.peopleAdmin[id] || { active: true, blocked: false, deleted: false };
    const peopleAdmin = { ...this.state.peopleAdmin, [id]: { ...current, ...patch } };
    try {
      localStorage.setItem('hub-people-admin', JSON.stringify(peopleAdmin));
    } catch (error) {
      this.notify('Could not save this account change. ' + (error instanceof Error ? error.message : String(error)), '#FF9B8A');
      return;
    }
    this.setState((st) => ({
      peopleAdmin,
      userResetTemp: ''
    }));
    return person;
  },
  togglePersonActive(id) {
    const current = this.state.peopleAdmin[id] || { active: true, blocked: false, deleted: false };
    const active = !current.active;
    const person = this.updatePeopleAdmin(id, { active });
    if (person) this.notify(person.name + (active ? ' was activated.' : ' was deactivated.'), '#B8F36B');
  },
  togglePersonBlocked(id) {
    const current = this.state.peopleAdmin[id] || { active: true, blocked: false, deleted: false };
    const blocked = !current.blocked;
    const person = this.updatePeopleAdmin(id, { blocked });
    if (person) this.notify(person.name + (blocked ? ' was blocked.' : ' was unblocked.'), '#B8F36B');
  },
  deletePerson(id) {
    const person = this.updatePeopleAdmin(id, { deleted: true, active: false });
    if (!person) return;
    this.setState({ userActionOpen: null, userResetTemp: '' });
    this.notify(person.name + ' was removed from the people directory.', '#FF9B8A');
  },
  async resetPersonPassword(id) {
    const person = this.dir.find((p) => p.id === id);
    if (!person) {
      this.notify('That person is no longer in the directory.', '#FF9B8A');
      this.setState({ userActionOpen: null, userResetTemp: '' });
      return;
    }
    const email = person.email;
    const random = new Uint8Array(9);
    crypto.getRandomValues(random);
    const temporaryPassword = 'Hub-' + Array.from(random, (n) => n.toString(36).padStart(2, '0')).join('') + '!';
    try {
      await useAccounts().resetPassword(email, temporaryPassword);
      this.setState({ userResetTemp: temporaryPassword });
      this.notify('Temporary password generated for ' + person.name + '.', '#B8F36B');
    } catch (error) {
      this.notify('Could not reset ' + person.name + '\'s password. ' + (error instanceof Error ? error.message : String(error)), '#FF9B8A');
    }
  }
};
