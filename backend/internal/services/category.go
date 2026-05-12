package services

import "strings"

func NormalizeMaintenanceCategory(value string) string {
	key := strings.ToUpper(strings.TrimSpace(value))
	key = strings.NewReplacer(",", " ", "-", " ", "_", " ", "/", " ").Replace(key)
	key = strings.Join(strings.Fields(key), " ")

	switch {
	case key == "PM SERVICE":
		return "PM Service"
	case key == "OIL CHANGE":
		return "Oil change"
	case key == "TIRE ISSUE":
		return "Tire issue"
	case key == "ENGINE ISSUE":
		return "Engine issue"
	case key == "TOWING":
		return "Towing"
	case key == "ROAD SERVICE":
		return "Road Service"
	case key == "BODY WORK":
		return "Body work"
	case key == "LEAKAGE":
		return "Leakage"
	case key == "KRIS SHOP" || strings.Contains(key, "JAX ALIGNMENT"):
		return "Kris Shop"
	case key == "TRUCK WASH DETAILING":
		return "Truck Wash/Detailing"
	case key == "ELECTRICAL ISSUE":
		return "Electrical issue"
	case key == "FLUIDS TRUCK PARTS":
		return "Fluids/Truck Parts"
	case key == "BRAKES DRUMS ROTORS":
		return "Brakes/Drums/Rotors"
	case key == "SCALE":
		return "Scale"
	case key == "OTHER":
		return "Other"
	case strings.Contains(key, "PM"):
		return "PM Service"
	case strings.Contains(key, "OIL"):
		return "Oil change"
	case strings.Contains(key, "TIRE"):
		return "Tire issue"
	case strings.Contains(key, "ENGINE"):
		return "Engine issue"
	case strings.Contains(key, "TOW"):
		return "Towing"
	case strings.Contains(key, "ROAD"):
		return "Road Service"
	case strings.Contains(key, "BODY"):
		return "Body work"
	case strings.Contains(key, "LEAK"):
		return "Leakage"
	case strings.Contains(key, "KRIS"):
		return "Kris Shop"
	case strings.Contains(key, "WASH") || strings.Contains(key, "DETAIL"):
		return "Truck Wash/Detailing"
	case strings.Contains(key, "ELECTR"):
		return "Electrical issue"
	case strings.Contains(key, "FLUID") || strings.Contains(key, "PART"):
		return "Fluids/Truck Parts"
	case strings.Contains(key, "BRAKE"):
		return "Brakes/Drums/Rotors"
	case strings.Contains(key, "SCALE"):
		return "Scale"
	default:
		return "Other"
	}
}
