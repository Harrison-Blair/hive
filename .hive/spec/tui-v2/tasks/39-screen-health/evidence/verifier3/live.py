import socket,json,os,subprocess,time,pathlib
root=pathlib.Path.cwd();d=root/'.hive/tmp/verify39v3'; pane='wA4:p2'
def call(method,params):
 s=socket.socket(socket.AF_UNIX);s.settimeout(10);s.connect(os.environ['HERDR_SOCKET_PATH']);s.sendall((json.dumps(dict(id='verify39',method=method,params=params))+'\n').encode());r=json.loads(s.makefile().readline());s.close();assert 'error' not in r,r;return r
assert call('ping',{})['result']['protocol']==22
subprocess.run(['herdr','pane','send-keys',pane,'ctrl+c','ctrl+c'],check=True);time.sleep(.5)
for board,size in [('Health','120x40'),('HealthAgents','120x40'),('CompactHealth','80x24')]:
 cmd=f"HIVE_HOME='{d}/home' COLORTERM=truecolor '{d}/hive' --demo '{root}/.hive/spec/tui-v2/design/{board}.fixture.json' --size {size}"
 subprocess.run(['herdr','pane','run',pane,cmd],check=True);time.sleep(1)
 for _ in range(10):
  a=call('pane.read',dict(pane_id=pane,source='visible',format='ansi'));time.sleep(.5);b=call('pane.read',dict(pane_id=pane,source='visible',format='ansi'))
  if a['result']['read']['text']==b['result']['read']['text']:break
 else:raise Exception('unstable')
 assert not a['result']['read']['truncated']
 meta=call('pane.get',dict(pane_id=pane));assert meta['result']['pane']['scroll']['offset_from_bottom']==0
 (d/(board+'.read.json')).write_text(json.dumps(a,indent=2));(d/(board+'.metadata.json')).write_text(json.dumps(meta,indent=2));(d/(board+'.ansi')).write_text(a['result']['read']['text']);subprocess.run(['go','run',str(d/'compare.go'),board],check=True)
 subprocess.run(['herdr','pane','send-keys',pane,'ctrl+c','ctrl+c'],check=True);time.sleep(.5)
