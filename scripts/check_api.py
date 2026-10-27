"""Real HTTP checks; no external Python dependencies. Run with no active draft."""
import json,urllib.request,urllib.error,uuid,pathlib,concurrent.futures
BASE='http://localhost:8085/api';checks=0
keys={'id','name','description','consumption_rate','fuel_tank_volume','image_url','video_url','creator','created_at','formed_at','is_owner','liked','likes_count'}
def check(condition,label):
 global checks
 assert condition,label
 checks+=1;print(f'PASS {checks:02d} {label}')
def call(method,path,body=None,raw=None,headers=None):
 if body is not None:raw=json.dumps(body).encode();headers={'Content-Type':'application/json'}
 req=urllib.request.Request(BASE+path,data=raw,method=method,headers=headers or {})
 try:r=urllib.request.urlopen(req)
 except urllib.error.HTTPError as e:r=e
 data=r.read();return r.status,json.loads(data) if data else None,data
p='/fuel_consumptions'
s,d,b=call('GET',p+'/draft');check(s==404 and b==b'','no draft initially')
s,d,b=call('GET',p+'?consumption_min=7&consumption_max=12');check(s==200 and len(d)>=2,'filter');check(all(7<=x['consumption_rate']<=12 for x in d),'inclusive numeric filter');check(all(set(x)==keys and 'status' not in x for x in d),'stable card keys, no status')
for query in ['consumption_min=x','consumption_min=20&consumption_max=1','consumption_max=NaN']:
 s,d,b=call('GET',p+'?'+query);check(s==404 and not b,'invalid filter empty404')
s,d,b=call('GET',p+'/feed');check(s==200 and d['id']==1,'first feed')
s,d,b=call('GET',p+'/feed?id=1&next=true');check(s==200 and d['id']==2,'next feed')
s,d,b=call('GET',p+'/feed?id=999999');check(s==404 and not b,'missing feed')
s,d,b=call('DELETE',p+'/5');check(s==404,'cannot delete other creator')
boundary='Lab3'+uuid.uuid4().hex
parts=[]
parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="name"\r\n\r\nПроверка ЛР3\r\n'.encode())
media=pathlib.Path(__file__).resolve().parents[1]/'media'
for field,ext,ct in [('image','jpg','image/jpeg'),('video','mp4','video/mp4')]:
 parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="{field}"; filename="car.{ext}"\r\nContent-Type: {ct}\r\n\r\n'.encode()+(media/f'cars_sedan.{ext}').read_bytes()+b'\r\n')
parts.append(f'--{boundary}--\r\n'.encode());raw=b''.join(parts);headers={'Content-Type':f'multipart/form-data; boundary={boundary}'}
s,car,b=call('POST',p,raw=raw,headers=headers);check(s==201,'create multipart draft');i=car['id'];check(set(car)==keys,'draft stable keys');check(car['consumption_rate'] is None and car['fuel_tank_volume'] is None and car['formed_at'] is None,'empty draft numeric fields retained');check(car['is_owner']==1,'owner flag')
for field,ext in [('image_url','jpg'),('video_url','mp4')]:check(urllib.request.urlopen(car[field]).read()==(media/f'cars_sedan.{ext}').read_bytes(),'actual MinIO bytes '+ext)
s,d,b=call('GET',p+'/draft');check(s==200 and d['id']==i,'draft without ID')
s,d,b=call('GET',p+f'/feed?id={i}');check(s==404,'draft hidden from feed')
s,d,b=call('POST',p,raw=raw,headers=headers);check(s==404,'one draft constraint')
s,d,b=call('PUT',p+f'/{i}/publication',{'status':'published'});check(s==404 and not b,'system fields rejected')
s,d,b=call('PUT',p+f'/{i}/publication',{'description':'','consumption_rate':8.4,'fuel_tank_volume':60});check(s==404,'publication requires fields')
s,d,b=call('PUT',p+f'/{i}/publication',{'description':'Дальняя поездка','consumption_rate':8.4,'fuel_tank_volume':60});check(s==200 and d['formed_at'] is not None,'publish sets date');check(set(d)==keys,'same keys after publication')
s,d,b=call('GET',p+'/draft');check(s==404,'published card leaves draft')
s,d,b=call('PUT',p+f'/{i}/publication',{'description':'Повтор','consumption_rate':8.4,'fuel_tank_volume':60});check(s==404,'no repeated publication')
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as ex:results=list(ex.map(lambda _:call('POST',p+f'/{i}/likes',{'liked':1}),range(4)))
check(all(r[0]==200 for r in results),'concurrent likes succeed')
s,d,b=call('GET',p+f'/feed?id={i}');check(d['likes_count']==1 and d['liked']==1,'likes unique')
for _ in range(2):
 s,d,b=call('POST',p+f'/{i}/likes',{'liked':0});check(s==200 and d['likes_count']==0 and d['liked']==0,'unlike idempotent')
s,d,b=call('POST',p+f'/{i}/likes',{'liked':2});check(s==404 and not b,'invalid like')
s,d,b=call('DELETE',p+f'/{i}');check(s==200 and not b,'soft delete empty200')
s,d,b=call('GET',p+f'/feed?id={i}');check(s==404 and not b,'deleted hidden')
s,d,b=call('POST',p+f'/{i}/likes',{'liked':1});check(s==404,'deleted cannot be liked')
login='test_'+uuid.uuid4().hex[:12]
s,d,b=call('POST','/users',{'login':login,'full_name':'Тест','password':'Lab3Demo123'});check(s==201 and set(d)=={'id','login','full_name'},'register without password/hash')
s,d,b=call('POST','/users',{'login':login,'full_name':'Тест','password':'Lab3Demo123'});check(s==404 and not b,'duplicate registration')
for url in ['/users/login','/users/logout']:
 s,d,b=call('POST',url);check(s==200 and not b,'auth stub')
print(f'All {checks} HTTP checks passed')
