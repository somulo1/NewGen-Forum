const base='http://localhost:8080/v1';
const raw=async(m,p,b,t)=>{const r=await fetch(base+p,{method:m,headers:{'Content-Type':'application/json',...(t?{Authorization:'Bearer '+t}:{})},body:b?JSON.stringify(b):undefined});const x=await r.text();if(!r.ok)throw new Error(r.status+' '+x);return x?JSON.parse(x):null;};
const t=await raw('POST','/auth/login',{email:'samwel.omulo@company.com',password:'Hub!Demo2026'});
const req=(m,p,b)=>raw(m,p,b,t.access_token);
const me=await req('GET','/people/'+t.org.person); console.log('me',JSON.stringify(me.reports_to));
const l=await req('GET','/people?limit=40');
for (const n of ['Amina Wanjiru','Grace Njeri','Brian Kiprop','David Otieno']) { const p=l.people.find(x=>x.name===n); if(p){const d=await req('GET','/people/'+p.id); console.log(n,d.role_name,JSON.stringify(d.reports_to));} }
