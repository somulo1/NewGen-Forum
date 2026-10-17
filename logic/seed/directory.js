// Sample data derived on start-up: directory. Called with the HubLogic instance as `this`.
export function seedDirectory() {
  // People directory (generated sample)
  const fn = ['Amina', 'Brian', 'Grace', 'David', 'Faith', 'Kevin', 'Mercy', 'John', 'Lilian', 'Peter', 'Esther', 'Daniel', 'Joy', 'Samuel', 'Ruth', 'Paul', 'Cynthia', 'Collins', 'Diana', 'Victor', 'Sharon', 'Dennis', 'Purity', 'George', 'Agnes', 'Felix', 'Irene', 'Moses', 'Naomi', 'Patrick', 'Rose', 'Stephen', 'Tabitha', 'Hassan', 'Fatuma', 'Ali', 'Zawadi', 'Kelvin', 'Beatrice', 'Allan'];
  const ln = ['Achieng', 'Wanjiru', 'Otieno', 'Kiprop', 'Njeri', 'Mutua', 'Wambui', 'Kamau', 'Chebet', 'Odhiambo', 'Akinyi', 'Mwangi', 'Nduta', 'Kibet', 'Atieno', 'Muthoni', 'Ochieng', 'Jeptoo', 'Wekesa', 'Nyambura', 'Omondi', 'Kariuki', 'Rotich', 'Barasa', 'Langat', 'Onyango', 'Gitau', 'Korir', 'Wafula', 'Maina'];
  const depts = [['Customer Care', ['Customer Care Agent', 'Team Lead, Customer Care']], ['Sales', ['Sales Executive', 'Relationship Manager', 'Sales Lead']], ['Finance', ['Accountant', 'Finance Analyst', 'Credit Officer']], ['Operations', ['Operations Officer', 'Branch Supervisor']], ['IT Operations', ['Systems Engineer', 'IT Support Officer', 'Network Engineer']], ['Software Development', ['Developer', 'QA Engineer', 'Product Designer']], ['HR', ['HR Officer', 'Recruiter']], ['Risk & Compliance', ['Risk Officer', 'Compliance Analyst', 'Internal Auditor']], ['Marketing', ['Marketing Officer', 'Content Creator']], ['Logistics', ['Driver', 'Logistics Coordinator']], ['Legal', ['Legal Officer']]];
  const locs = ['Head Office', 'North Branch', 'Coast Branch', 'West Branch'];
  const cols = ['#7CC4FF', '#FFB86B', '#F49AC1', '#6EE7C8', '#B8F36B', '#C9B6FF', '#FFD66B'];
  const sts = ['Available', 'Available', 'In a meeting', 'Available', 'Focusing', 'Away', 'Available'];
  this.dir = [];
  for (let i = 0; i < 248; i++) {
    const d = depts[(i * 7 + (i >> 2)) % depts.length];
    const first = fn[(i * 13) % fn.length], last = ln[(i * 17 + 3) % ln.length];
    const loc = locs[(i * 5 + (i >> 3)) % (i % 3 === 0 ? 1 : 4)];
    const br = this.org.branches.find((b) => b.name === loc);
    const sec = br ? this.org.sectors.find((x) => x.id === br.sector) : null;
    const us = br ? this.org.units.filter((u) => u.branch === br.id) : [];
    const unit = us.length ? (i % (us.length + 1) === us.length ? '' : us[i % (us.length + 1)].name) : '';
    const name = first + ' ' + last;
    this.dir.push({ id: i, name, email: name.toLowerCase().replace(/ /g, '.') + '.' + (i + 1) + '@company.com', initials: first[0] + last[0], dept: d[0], title: d[1][i % d[1].length], loc: loc, sector: sec ? sec.name : '', unit: unit, color: cols[i % cols.length], status: sts[(i * 3) % sts.length] });
  }

  // Named colleagues are part of the directory too
  [['Brian Kiprop', 'IT Operations', 'Systems Engineer', 'Head Office', '#7CC4FF'], ['David Otieno', 'Software Development', 'Developer', 'Head Office', '#F49AC1'], ['Amina Wanjiru', 'Finance', 'Finance Analyst', 'Head Office', '#6EE7C8'],
    ['Grace Njeri', 'HR', 'HR Officer', 'Head Office', '#FFB86B'], ['Faith Achieng', 'Customer Care', 'Customer Care Agent', 'Coast Branch', '#B8F36B'], ['Kevin Mutua', 'Sales', 'Sales Lead', 'North Branch', '#C9B6FF'],
    ['Mercy Wambui', 'Finance', 'Accountant', 'Head Office', '#FFB86B'], ['John Kamau', 'Logistics', 'Driver', 'West Branch', '#7CC4FF'], ['Lilian Chebet', 'Marketing', 'Marketing Officer', 'Head Office', '#F49AC1']].forEach(([name, dept, title, loc, color]) => {
    const br = this.org.branches.find((b) => b.name === loc);
    const sec = br ? this.org.sectors.find((x) => x.id === br.sector) : null;
    this.dir.push({ id: this.dir.length, name: name, email: name.toLowerCase().replace(/ /g, '.') + '.' + (this.dir.length + 1) + '@company.com', initials: name.split(' ').map((x) => x[0]).join(''), dept: dept, title: title, loc: loc, sector: sec ? sec.name : '', unit: '', color: color, status: 'Available' });
  });
}
