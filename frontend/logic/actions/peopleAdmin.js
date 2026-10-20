import * as peopleW from '~/logic/server/widgets/people.js';

// Connected to the server, an account's status (active, inactive, blocked, deleted) and password resets are the server's: they
// are saved in the database, take effect at once on every device, and are checked against the rights of whoever does them.
async function serverStatus(app, id, action, done) {
  const person = app.dir.find((p) => p.id === id);
  if (!person) { app.notify('That person is no longer in the directory.', '#FF9B8A'); app.setState({ userActionOpen: null, userResetTemp: '' }); return; }
  try {
    await app.srv.request('POST', `/people/${person.uid}/status`, { action });
    await peopleW.load(app, app.srv);
    app.setState({ userActionOpen: null, userResetTemp: '' });
    app.notify(done(person), action === 'delete' || action === 'block' ? '#FF9B8A' : '#B8F36B');
  } catch (e) { app.notify((e && e.message) || 'That did not work.', '#FF9B8A'); }
}

export const peopleAdminActions = {
  togglePersonActive(id) {
    const p = this.dir.find((x) => x.id === id); const on = !!p && p.accStatus === 'active';
    return serverStatus(this, id, on ? 'deactivate' : 'activate', (x) => x.name + (on ? ' was deactivated.' : ' was activated.'));
  },
  togglePersonBlocked(id) {
    const p = this.dir.find((x) => x.id === id); const on = !!p && p.accStatus === 'blocked';
    return serverStatus(this, id, on ? 'unblock' : 'block', (x) => x.name + (on ? ' was unblocked.' : ' was blocked.'));
  },
  deletePerson(id) {
    return serverStatus(this, id, 'delete', (x) => x.name + ' was removed from the people directory.');
  },
  async resetPersonPassword(id) {
    const person = this.dir.find((p) => p.id === id);
    if (!person) {
      this.notify('That person is no longer in the directory.', '#FF9B8A');
      this.setState({ userActionOpen: null, userResetTemp: '' });
      return;
    }
    try {
      const r = await this.srv.request('POST', `/people/${person.uid}/reset-password`, {});
      this.setState({ userResetTemp: r.temporary_password });
      this.notify('Temporary password generated for ' + person.name + '.', '#B8F36B');
    } catch (e) { this.notify((e && e.message) || 'Could not reset the password.', '#FF9B8A'); }
  }
};
