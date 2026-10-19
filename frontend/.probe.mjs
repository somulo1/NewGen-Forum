import { Device, b64 } from './logic/server/e2ee.js';
import { Vault } from './logic/server/vault.js';
const base='http://localhost:8080/v1';
const raw=async(m,p,b,t)=>{const r=await fetch(base+p,{method:m,headers:{'Content-Type':'application/json',...(t?{Authorization:'Bearer '+t}:{})},body:b?JSON.stringify(b):undefined});const x=await r.text();if(!r.ok)throw new Error(r.status+' '+x);return x?JSON.parse(x):null;};
const email='samwel.omulo@company.com';
const t=await raw('POST','/auth/login',{email,password:'Hub!Demo2026',org:'demo'});
const tok=t.access_token; const req=(m,p,b)=>raw(m,p,b,tok);
const devs=(await req('GET','/e2ee/devices')).devices; const dev=Device.fromDemo(email); const mine=devs.find(d=>d.identity_key===b64(dev.identity));
const vault=new Vault(req,{org:t.org.id,account:t.account,person:t.org.person,device:dev,deviceId:mine.id});
globalThis.__v={req,vault,t};
const {conversations}=await req('GET','/conversations?kind=forum&limit=100');
console.log(JSON.stringify(conversations.map(c=>({id:c.id,name:c.name,about:c.about,privacy:c.privacy,my_role:c.my_role,unread:c.unread,mc:c.member_count,last:c.last_at})),null,1));
const f=conversations[0];
const full=await req('GET','/conversations/'+f.id); console.log(JSON.stringify({...full,members:(full.members||[]).slice(0,2)}));
const ms=(await req('GET','/conversations/'+f.id+'/messages?limit=5')).messages;
for(const m of ms){console.log(m.type,m.reply_to,m.sender_name,m.created_at,JSON.stringify(await vault.openMessage(m)));}
console.log(Object.keys(await req('GET','/conversations/'+f.id+'/messages?limit=5')));
