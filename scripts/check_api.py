"""Exercise the real HTTP API, PostgreSQL and MinIO. Run on the Lab 3 demo database."""
import json, os, urllib.request, urllib.error, uuid, concurrent.futures
from pathlib import Path
BASE=os.getenv('API_URL','http://localhost:8085')+'/api'
checks=0
def check(value,label):
 global checks
 assert value,label
 checks+=1
 print('PASS',label)
def call(method,path,body=None,code=200):
 data=json.dumps(body).encode() if body is not None else None
 req=urllib.request.Request(BASE+path,data=data,method=method,headers={'Content-Type':'application/json'})
 try:r=urllib.request.urlopen(req,timeout=20)
 except urllib.error.HTTPError as e:r=e
 raw=r.read();check(r.status==code,f'{method} {path}: {r.status}')
 if code>=400:check(raw==b'','error body is empty')
 return json.loads(raw) if raw else None
cars=call('GET','/cars');check(len(cars)>=4,'seed cars present')
check('status' not in cars[0],'service status hidden')
check(all(7<=x['consumption_rate']<=8 for x in call('GET','/cars?consumption_min=7&consumption_max=8')),'range filter')
check(call('GET','/cars?consumption_min=900')==[],'empty list is array')
call('GET','/cars?consumption_min=20&consumption_max=10',code=404)
call('GET','/cars?consumption_min=NaN',code=404)
call('GET','/cars/999999',code=404)
call('POST','/fuel_consumptions',{},404)
cart=call('GET','/fuel_consumptions/cart')
check(cart['id'] is None,'start with empty draft; delete existing draft before rerunning')
r=call('POST','/fuel_consumptions/draft/cars/1',code=201);rid=r['id'];keys=set(r);itemkeys=set(r['items'][0]);check(r['total_saved_liters'] is None,'calculation is null before formation')
check(r['saving_percent'] is None and r['moderator'] is None,'nullable fields present')
call('POST','/fuel_consumptions/draft/cars/1',code=404)
r=call('POST','/fuel_consumptions/draft/cars/2',code=201);check(r['id']==rid and len(r['items'])==2,'same draft contains two cars')
check(call('GET','/fuel_consumptions/cart')['count']==2,'cart counts cars')
check(not any(x['id']==rid for x in call('GET','/fuel_consumptions')),'draft excluded from list')
call('PUT',f'/fuel_consumptions/{rid}',{'creator_id':2},404)
call('PUT',f'/fuel_consumptions/{rid}/completion',{'decision':'complete'},404)
call('PUT',f'/fuel_consumptions/{rid}/formation',code=404)
call('PUT',f'/fuel_consumptions/{rid}/cars/1',{'distance_km':500})
call('PUT',f'/fuel_consumptions/{rid}/cars/2',{'distance_km':500})
call('PUT',f'/fuel_consumptions/{rid}',{'route':'Москва — Тверь — Москва','saving_percent':10})
r=call('PUT',f'/fuel_consumptions/{rid}/formation');check(set(r)==keys and set(r['items'][0])==itemkeys,'stable JSON keys')
check(r['total_saved_liters']==8.85 and r['calculated_count']==2,'expected total 8.85 L')
check(r['items'][0]['range_km']==771.6,'tank based range 771.6 km')
check(r['status']=='formed' and r['formed_at'] is not None,'formation timestamp')
check(call('GET','/fuel_consumptions/cart')['id'] is None,'cart cleared on formation')
call('PUT',f'/fuel_consumptions/{rid}',{'route':'changed'},404)
call('DELETE',f'/fuel_consumptions/{rid}',code=404)
call('PUT',f'/fuel_consumptions/{rid}/cars/1',{'distance_km':10},404)
r=call('PUT',f'/fuel_consumptions/{rid}/completion',{'decision':'complete'});check(r['moderator']=='moderator' and r['completed_at'] is not None,'completion assigned server moderator')
date=r['formed_at'][:10]
check(any(x['id']==rid for x in call('GET',f'/fuel_consumptions?status=completed&formed_from={date}&formed_to={date}')),'date range inclusive of end day')
call('PUT',f'/fuel_consumptions/{rid}/completion',{'decision':'reject'},404)
# Concurrent adds produce one draft, safeguarded by a row lock and partial unique index.
with concurrent.futures.ThreadPoolExecutor(2) as pool:
 rows=list(pool.map(lambda i:call('POST',f'/fuel_consumptions/draft/cars/{i}',code=201),[1,2]))
check(rows[0]['id']==rows[1]['id'],'concurrent requests share a single draft')
rid2=rows[0]['id'];call('DELETE',f'/fuel_consumptions/{rid2}/cars/1')
check(call('GET','/fuel_consumptions/cart')['count']==1,'remove by composite pair')
call('DELETE',f'/fuel_consumptions/{rid2}')
call('GET',f'/fuel_consumptions/{rid2}',code=404)
# Reject branch.
r=call('POST','/fuel_consumptions/draft/cars/1',code=201);ri=r['id']
call('PUT',f'/fuel_consumptions/{ri}/cars/1',{'distance_km':100})
call('PUT',f'/fuel_consumptions/{ri}',{'route':'Проверка отклонения','saving_percent':0})
call('PUT',f'/fuel_consumptions/{ri}/formation')
check(call('PUT',f'/fuel_consumptions/{ri}/completion',{'decision':'reject'})['status']=='rejected','rejection branch')
login='test_'+uuid.uuid4().hex[:10]
u=call('POST','/users',{'login':login,'full_name':'Тестовый водитель','password':'DemoPass123'},201)
check(set(u)=={'id','login','full_name'},'password never returned')
call('POST','/users',{'login':login,'full_name':'Тест','password':'DemoPass123'},404)
call('POST','/users/login',{});call('POST','/users/logout',{})
# Actual multipart upload and verification from MinIO.
boundary=uuid.uuid4().hex; parts=[]
for k,v in {'name':'Тестовый автомобиль','description':'','consumption_rate':'8.1','fuel_tank_volume':'55'}.items():
 parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="{k}"\r\n\r\n{v}\r\n'.encode())
root=Path(__file__).resolve().parents[1]
for k,ext,ct in [('image','jpg','image/jpeg'),('video','mp4','video/mp4')]:
 parts.append(f'--{boundary}\r\nContent-Disposition: form-data; name="{k}"; filename="пример.{ext}"\r\nContent-Type: {ct}\r\n\r\n'.encode()+ (root/f'media/cars_sedan.{ext}').read_bytes()+b'\r\n')
data=b''.join(parts)+f'--{boundary}--\r\n'.encode()
req=urllib.request.Request(BASE+'/cars',data=data,headers={'Content-Type':f'multipart/form-data; boundary={boundary}'},method='POST')
with urllib.request.urlopen(req,timeout=60) as res:
 check(res.status==201,'multipart create');car=json.load(res)
for k,ext in [('image_url','jpg'),('video_url','mp4')]:
 with urllib.request.urlopen(car[k],timeout=20) as res:
  check(res.read()==(root/f'media/cars_sedan.{ext}').read_bytes(),f'{k} stored intact in MinIO')
check(car['description']=='' and 'status' not in car,'empty description included; internal status hidden')
print(f'TOTAL {checks} checks passed')
