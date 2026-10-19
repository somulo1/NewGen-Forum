import { Device, b64 } from './logic/server/e2ee.js';
import { Vault } from './logic/server/vault.js';
const base='http://localhost:8080/v1';
const raw=async(m,p,b,t)=>{const r=await fetch(base+p,{method:m,headers:{'Content-Type':'application/json',...(t?{Authorization:'Bearer '+t}:{})},body:b?JSON.stringify(b):undefined});const x=await r.text();if(!r.ok)throw new Error(r.status+' '+x);return x?JSON.parse(x):null;};
const email='samwel.omulo@company.com';
const t=await raw('POST','/auth/login',{email,password:'Hub!Demo2026'});
const tok=t.access_token; const req=(m,p,b)=>raw(m,p,b,tok);
const devs=(await req('GET','/e2ee/devices')).devices; const dev=Device.fromDemo(email); const mine=devs.find(d=>d.identity_key===b64(dev.identity));
const vault=new Vault(req,{org:t.org.id,account:t.account,person:t.org.person,device:dev,deviceId:mine.id});
const cmd=process.argv[2];
const {conversations}=await req('GET','/conversations?kind=forum&limit=100');
const it=conversations.find(c=>c.name==='IT Operations');
if(cmd==='post'){const s=await vault.seal(it.id,{text:process.argv[3]});await req('POST','/conversations/'+it.id+'/messages',{...s,type:'text'});console.log('posted');}
if(cmd==='list'){const ms=(await req('GET','/conversations/'+it.id+'/messages?limit=6')).messages;for(const m of ms)console.log(m.type,m.deleted?'DELETED':'',m.reply_to?'reply':'',JSON.stringify(await vault.openMessage(m)));}
if(cmd==='clean'){const ms=(await req('GET','/conversations/'+it.id+'/messages?limit=30')).messages;for(const m of ms){const pl=m.deleted?null:await vault.openMessage(m);if(pl&&(/UI test|Realtime test|Reply test/.test(pl.text||'')||(pl.file&&pl.file[0].name==='File test.txt'))){await req('DELETE','/conversations/'+it.id+'/messages/'+m.id);console.log('deleted',pl.text);}}}
