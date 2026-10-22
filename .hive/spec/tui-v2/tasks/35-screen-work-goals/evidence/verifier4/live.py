import importlib.util,subprocess,time,pathlib,shlex,json
root=pathlib.Path.cwd()
spec=importlib.util.spec_from_file_location('capture',root/'.hive/tmp/verify35v4/capture.py'); c=importlib.util.module_from_spec(spec);spec.loader.exec_module(c)
def key(name,*keys,w=120,h=40):
 subprocess.run(['herdr','pane','send-keys','wA2:p2',*keys],check=True,stdout=subprocess.DEVNULL);time.sleep(.2);c.capture(name,w,h)
def quit():
 subprocess.run(['herdr','pane','send-keys','wA2:p2','ctrl+c','ctrl+c'],check=True);time.sleep(.7)
def start(w,h):
 cmd='env HIVE_HOME='+shlex.quote(str(root/'.hive/tmp/verify35v4/home'))+' HIVE_NO_STARTUP=1 COLORTERM=truecolor '+shlex.quote(str(root/'.hive/tmp/verify35v4/hive'))+' --demo '+shlex.quote(str(root/'.hive/spec/tui-v2/design/WorkGoals.fixture.json'))+f' --size {w}x{h}'
 subprocess.run(['herdr','pane','run','wA2:p2',cmd],check=True);time.sleep(.7)
start(120,40)
c.capture('start')
for name,keys in [('s',['s']),('up',['up']),('j',['j']),('k',['k']),('down',['down']),('g',['g']),('G',['G']),('details',['tab']),('details-scroll',['down']),('list',['esc']),('design',['o']),('task',['t']),('queue',[']']),('back',['[']),('help',['?']),('help-close',['esc']),('refresh',['R'])]: key(name,*keys)
quit()
for w,h in [(119,40),(80,24),(79,24)]:
 start(w,h);c.capture(f'c{w}',w,h)
 if w>=80:
  key(f'c{w}-enter','enter',w=w,h=h)
  key(f'c{w}-down','down',w=w,h=h)
  key(f'c{w}-esc','esc',w=w,h=h)
  key(f'c{w}-k','k',w=w,h=h)
  key(f'c{w}-s','s',w=w,h=h)
 quit()
