import subprocess as sp, pathlib, time, json, socket, os, re
root=pathlib.Path.cwd(); scratch=root/'.hive/tmp/t38-verifier'; out=root/'.hive/spec/tui-v2/tasks/38-screen-friction/evidence/verifier'; design=root/'.hive/spec/tui-v2/design'; pane='wA1:p2'
def cli(*args): return sp.check_output(['herdr',*args]).decode()
def rpc(method,params):
 with socket.socket(socket.AF_UNIX) as s:
  s.connect(os.environ['HERDR_SOCKET_PATH']); s.sendall((json.dumps(dict(id='verify38',method=method,params=params))+'\n').encode()); return json.loads(s.makefile().readline())
def cap(name):
 prev=None
 for _ in range(20):
  reply=rpc('pane.read',dict(pane_id=pane,source='visible',format='ansi')); read=reply['result']['read']; text=read['text']; assert not read['truncated']
  if text==prev: break
  prev=text; time.sleep(.5)
 else: raise Exception('not stable')
 info=json.loads(cli('pane','get',pane)); assert info['result']['pane']['scroll']['offset_from_bottom']==0
 (out/(name+'.json')).write_text(json.dumps(dict(read=reply,pane=info)))
 (out/(name+'.ansi')).write_text(text)
 return re.sub(r'\x1b\[[0-9;]*m','',text)
def press(k):
 if k in ('pgdown','pgup'): rpc('pane.send_text',dict(pane_id=pane,text='\x1b[6~' if k=='pgdown' else '\x1b[5~'))
 else: cli('pane','send-keys',pane,k)
 time.sleep(.25)
def start(board,w,h):
 cmd=f'clear; HIVE_HOME={scratch}/home COLORTERM=truecolor {scratch}/hive --demo {design}/{board}.fixture.json --size {w}x{h}'
 cli('pane','run',pane,cmd);time.sleep(1)
def stop(): press('ctrl+c');press('ctrl+c');time.sleep(.3)
# A prior setup run may still be active.
stop()
observations=[]
for board,w,h in [('Friction',120,40),('FrictionResolved',120,40),('CompactFriction',80,24)]:
 start(board,w,h);text=cap(board)
 result=sp.check_output([str(scratch/'compare'),str(design/(board+'.grid')),str(out/(board+'.ansi')),str(w),str(h)]).decode();(out/(board+'-diff.txt')).write_text(result); assert result.rstrip().endswith('differences under the live rule 0'), result
 if board=='Friction':
  for k in ['j','down','k','up','G','g','pgdown','pgup',']','[','tab','down','G','g','esc','enter','j','G','esc','a','/','?','esc']:
   press(k); name=f'key-{len(observations):02}'; t=cap(name);observations.append(dict(board=board,key=k,capture=name,footer=t.splitlines()[39]))
 elif board=='CompactFriction':
  for k in ['enter','1','down','pgdown','esc',']','[','?','esc']:
   press(k);name=f'key-{len(observations):02}';t=cap(name);observations.append(dict(board=board,key=k,capture=name,footer=t.splitlines()[23]))
 stop()
(out/'keys.json').write_text(json.dumps(observations,indent=2))
print('All three live grids match. Key captures saved.')
