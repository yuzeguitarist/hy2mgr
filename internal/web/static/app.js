const el = (tag, attrs={}, children=[]) => {
  const n = document.createElement(tag);
  Object.entries(attrs).forEach(([k,v])=>{
    if(k==="class") n.className=v;
    else if(k==="html") n.innerHTML=v;
    else n.setAttribute(k,v);
  });
  children.forEach(c => n.appendChild(typeof c==="string"?document.createTextNode(c):c));
  return n;
};

async function api(path, opts={}){
  const r = await fetch(path, {credentials:'same-origin', ...opts});
  if(r.status===401){ location.href="/login"; return; }
  const ct = r.headers.get('content-type')||'';
  if(ct.includes('application/json')) return await r.json();
  return await r.text();
}

function nav(){
  return el('div',{class:'nav'},[
    el('a',{href:'#dashboard'},['Dashboard']),
    el('a',{href:'#nodes'},['Nodes']),
    el('a',{href:'#logs'},['Logs']),
    el('a',{href:'#settings'},['Settings']),
    el('a',{href:'/logout'},['Logout']),
  ]);
}

function card(title, body){
  return el('div',{class:'card'},[
    el('div',{class:'row'},[
      el('h2',{style:'margin:0;flex:1'},[title]),
    ]),
    body
  ]);
}

async function renderDashboard(root){
  const d = await api('/api/dashboard');
  const body = el('div',{},[
    el('div',{class:'row'},[
      el('span',{class:'badge'},['Hysteria: '+d.hysteriaStatus]),
      el('span',{class:'badge'},['Listen UDP: '+d.listen]),
      el('span',{class:'badge'},['PinSHA256: '+d.pin]),
    ]),
    el('p',{class:'small'},['Tip: cloud provider security group must allow UDP/'+d.port+'.']),
    el('h3',{},['Recent errors (journal)']),
    el('pre',{},[d.recentErrors||'(none)']),
  ]);
  root.appendChild(card('Service', body));
}

async function renderNodes(root){
  const ns = await api('/api/nodes');
  const sub = await api('/api/subscription');
  const body = el('div',{},[]);
  const actions = el('div',{class:'row'},[
    el('button',{class:'btn primary',id:'addNode'},['+ Add node']),
    el('a',{class:'btn',href:sub.url, target:'_blank'},['Open subscription URL']),
    el('button',{class:'btn',id:'rotateToken'},['Rotate subscription token']),
  ]);
  body.appendChild(actions);

  const t = el('table',{},[]);
  t.appendChild(el('tr',{},[
    el('th',{},['Name']),
    el('th',{},['Enabled']),
    el('th',{},['Actions']),
  ]));
  ns.nodes.forEach(n=>{
    const row = el('tr',{},[]);
    row.appendChild(el('td',{},[
      el('div',{},[n.name]),
      el('div',{class:'small'},['ID: '+n.id+'  User: '+n.username]),
    ]));
    row.appendChild(el('td',{},[n.enabled?'✅':'⛔']));
    const act = el('td',{},[]);
    const btnCopy = el('button',{class:'btn'},['Copy URI']);
    btnCopy.onclick=async()=>{
      const r = await api('/api/nodes/'+n.id+'/uri');
      await navigator.clipboard.writeText(r.uri);
      alert('Copied.');
    };
    const btnQR = el('button',{class:'btn'},['QR']);
    btnQR.onclick=()=>showQR(n.id);
    const btnDis = el('button',{class:'btn'},[n.enabled?'Disable':'Enable']);
    btnDis.onclick=async()=>{
      await api('/api/nodes/'+n.id+(n.enabled?'/disable':'/enable'), {method:'POST'});
      location.hash='#nodes'; route();
    };
    const btnReset = el('button',{class:'btn'},['Reset pass']);
    btnReset.onclick=async()=>{
      if(!confirm('Reset password? The old one will stop working.')) return;
      const r = await api('/api/nodes/'+n.id+'/reset', {method:'POST'});
      alert('New password generated. Copy new URI now.');
    };
    const btnDel = el('button',{class:'btn danger'},['Delete']);
    btnDel.onclick=async()=>{
      if(!confirm('Delete node?')) return;
      await api('/api/nodes/'+n.id, {method:'DELETE'});
      route();
    };
    act.appendChild(el('div',{class:'row'},[btnCopy, btnQR, btnDis, btnReset, btnDel]));
    row.appendChild(act);
    t.appendChild(row);
  });
  body.appendChild(t);

  // modal
  const modal = el('div',{id:'modal',style:'display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);align-items:center;justify-content:center'},[
    el('div',{class:'card',style:'max-width:680px;width:92vw'},[
      el('div',{class:'row'},[
        el('h3',{style:'margin:0;flex:1'},['Node QR']),
        el('button',{class:'btn',id:'closeModal'},['Close']),
      ]),
      el('div',{id:'modalBody'},[])
    ])
  ]);
  body.appendChild(modal);

  async function showQR(id){
    const uri = await api('/api/nodes/'+id+'/uri');
    const mb = document.getElementById('modalBody');
    mb.innerHTML='';
    mb.appendChild(el('p',{},['URI:']));
    mb.appendChild(el('pre',{},[uri.uri]));
    const img = el('img',{src:'/api/nodes/'+id+'/qrcode.png', style:'max-width:280px;border-radius:10px;border:1px solid #22335f'});
    mb.appendChild(img);
    mb.appendChild(el('div',{class:'row'},[
      el('a',{class:'btn',href:'/api/nodes/'+id+'/qrcode.png'},['Download PNG']),
      el('a',{class:'btn',href:'/api/nodes/'+id+'/qrcode.svg'},['Download SVG']),
    ]));
    modal.style.display='flex';
  }
  body.querySelector('#closeModal').onclick=()=>modal.style.display='none';

  body.querySelector('#addNode').onclick=async()=>{
    const name = prompt('Node name?','my-phone');
    if(!name) return;
    const r = await api('/api/nodes', {method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify({name})});
    alert('Node created. Copy URI from table.');
    route();
  };
  body.querySelector('#rotateToken').onclick=async()=>{
    if(!confirm('Rotate token? Old subscription URL will stop working.')) return;
    const r = await api('/api/subscription/rotate', {method:'POST'});
    alert('New token generated.');
    route();
  };

  root.appendChild(card('Nodes', body));
}

async function renderLogs(root){
  const t = await api('/api/logs?lines=200');
  root.appendChild(card('Logs (journalctl -u hysteria-server.service)', el('pre',{},[t])));
}

async function renderSettings(root){
  const s = await api('/api/settings');
  const body = el('div',{},[]);
  const form = el('div',{},[
    el('div',{class:'row'},[
      el('div',{},[
        el('label',{},['Default SNI']),
        el('input',{id:'sni',value:s.sni}),
      ]),
      el('div',{},[
        el('label',{},['Masquerade URL']),
        el('input',{id:'masq',value:s.masqueradeUrl}),
      ]),
    ]),
    el('div',{class:'row'},[
      el('div',{},[
        el('label',{},['Preferred UDP port']),
        el('input',{id:'port',value:s.listenPort, inputmode:'numeric'}),
      ]),
      el('div',{},[
        el('label',{},['Rewrite Host']),
        el('input',{id:'rewrite', type:'checkbox'}),
      ]),
    ]),
    el('div',{class:'row'},[
      el('button',{class:'btn primary',id:'save'},['Save & Apply']),
      el('button',{class:'btn',id:'rotateCert'},['Rotate cert']),
      el('button',{class:'btn',id:'setPass'},['Change admin password']),
    ]),
    el('p',{class:'small'},['Saving will regenerate /etc/hysteria/config.yaml and restart the service.']),
  ]);
  form.querySelector('#rewrite').checked = s.masqueradeRewrite;

  form.querySelector('#save').onclick=async()=>{
    const payload = {
      sni: form.querySelector('#sni').value.trim(),
      masqueradeUrl: form.querySelector('#masq').value.trim(),
      masqueradeRewrite: form.querySelector('#rewrite').checked,
      listenPort: parseInt(form.querySelector('#port').value,10),
    };
    const r = await api('/api/settings', {method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify(payload)});
    alert('Saved.');
    route();
  };

  form.querySelector('#rotateCert').onclick=async()=>{
    if(!confirm('Rotate self-signed cert? Clients should update pinSHA256.')) return;
    await api('/api/cert/rotate', {method:'POST'});
    alert('Rotated. See dashboard for new pin.');
    route();
  };

  form.querySelector('#setPass').onclick=async()=>{
    const p = prompt('New admin password (min 12 chars recommended):');
    if(!p) return;
    await api('/api/admin/password', {method:'POST', headers:{'content-type':'application/json'}, body: JSON.stringify({password:p})});
    alert('Password updated.');
  };

  body.appendChild(form);
  root.appendChild(card('Settings', body));
}

async function route(){
  const root = document.getElementById('app');
  root.innerHTML='';
  root.appendChild(nav());
  const cont = el('div',{class:'container'},[]);
  root.appendChild(cont);
  const h = (location.hash||'#dashboard').replace('#','');
  if(h==='dashboard') await renderDashboard(cont);
  else if(h==='nodes') await renderNodes(cont);
  else if(h==='logs') await renderLogs(cont);
  else if(h==='settings') await renderSettings(cont);
  else await renderDashboard(cont);
}

window.addEventListener('hashchange', route);
route();
