package service
import (
 "testing"
 "github.com/arseniy9606/cruisecontrol---back/internal/model"
)
func TestCalculation(t *testing.T){p,d:=10.0,500.0;r:=model.FuelConsumption{Route:"Москва — Тверь",SavingPercent:&p,Items:[]model.TripCar{{DistanceKM:&d,Car:model.Car{ConsumptionRate:7.2,FuelTankVolume:50,Status:"published"}}}};if e:=Calculate(&r);e!=nil{t.Fatal(e)};x:=r.Items[0];if *x.BaseLiters!=36||*x.CruiseLiters!=32.4||*x.SavedLiters!=3.6||*x.RangeKM!=771.6||*r.TotalSavedLiters!=3.6{t.Fatalf("unexpected result: %+v",x)}}
func TestInvalidCalculation(t *testing.T){p:=100.0;r:=model.FuelConsumption{Route:"test",SavingPercent:&p};if Calculate(&r)==nil{t.Fatal("invalid form accepted")}}
func TestSingleton(t *testing.T){if CurrentUser()!=CurrentUser()||CurrentUser().CreatorID!=1{t.Fatal("identity must be singleton")}}
