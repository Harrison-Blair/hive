import subprocess,json,time,pathlib,re,html,sys
out=pathlib.Path('.hive/spec/tui-v2/tasks/35-screen-work-goals/evidence/verifier3')
def call(*args):
    import socket,os
    method='.'.join(args[:2]); params={'pane_id':args[2]}
    if method=='pane.read': params.update(source='visible',format='ansi')
    with socket.socket(socket.AF_UNIX) as s:
        s.connect(os.environ['HERDR_SOCKET_PATH']); s.sendall((json.dumps({'id':'verify35','method':method,'params':params})+'\n').encode())
        return json.loads(s.makefile().readline())['result']
def capture(name,w=120,h=40):
    for _ in range(20):
        a=call('pane','read','w9N:p2','--source','visible','--format','ansi')['read']; time.sleep(.5)
        b=call('pane','read','w9N:p2','--source','visible','--format','ansi')['read']
        if a['text']==b['text']: break
    else: raise Exception('unstable capture')
    info=call('pane','get','w9N:p2')['pane']; assert not b['truncated'] and info['scroll']['offset_from_bottom']==0
    (out/(name+'.json')).write_text(json.dumps({'read':b,'pane':info},indent=2))
    text='\n'.join(b['text'].splitlines()[:h])
    # All board symbols are one cell. Retain SGR while cropping the left w cells.
    lines=[]
    for line in text.splitlines():
        parts=re.findall(r'\x1b\[[0-9;]*m|.',line); n=0; crop=''
        for p in parts:
            if p.startswith('\x1b'): crop+=p
            elif n<w: crop+=p;n+=1
            else: break
        lines.append(crop+'\x1b[0m')
    text='\n'.join(lines); (out/(name+'.ansi')).write_text(text)
    plain=re.sub(r'\x1b\[[0-9;]*m','',text); (out/(name+'.txt')).write_text(plain)
    # Render terminal SGR as HTML for the independent PNG capture.
    color='#ddd'; bold=False; spans=[]
    for p in re.split(r'(\x1b\[[0-9;]*m)',text):
        if p.startswith('\x1b'):
            vals=[int(x or 0) for x in p[2:-1].split(';')];i=0
            while i<len(vals):
                v=vals[i]
                if v==0: color='#ddd';bold=False
                elif v==1: bold=True
                elif v==22: bold=False
                elif v==39: color='#ddd'
                elif v==38 and vals[i+1]==2: color='#%02x%02x%02x'%tuple(vals[i+2:i+5]);i+=4
                i+=1
        else: spans.append('<span style="color:'+color+';font-weight:'+('bold' if bold else 'normal')+'">'+html.escape(p)+'</span>')
    (out/(name+'.html')).write_text('<body style="background:#171715;margin:0"><pre style="font:14px/18px monospace;margin:0">'+''.join(spans)+'</pre></body>')
    print(name,repr(plain.splitlines()[-1]))
if __name__=='__main__': capture(sys.argv[1],int(sys.argv[2]) if len(sys.argv)>2 else 120,int(sys.argv[3]) if len(sys.argv)>3 else 40)
