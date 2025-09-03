// Package seeder initial data for db
package seeder

import "github.com/google/uuid"

type AppFeature struct {
	ID               uuid.UUID
	Title            string
	Subtitle         string
	Name             string
	Details          []string
	Description      string
	QuickDescription string
}

type AppRoles struct {
	ID          uuid.UUID
	Name        string
	Permissions string
}

//  maybe make a lambda that can be called with strings to save to s3 for feature

func GetAppFeatures() []AppFeature {
	return []AppFeature{
		{
			ID:               uuid.New(),
			Name:             "Payments and Finances",
			Title:            "Streamline Payments & Finances",
			Subtitle:         "Automated Rent Collection",
			QuickDescription: "Collect rent automatically",
			Description:      "Set up recurring rent payments and get notified instantly when tenants pay",
			Details: []string{
				"Set up <strong>automated rent subscriptions</strong> with recurring payments.",
				"Keep track of <strong>property taxes</strong> and upcoming payment dates.",
				"View income and expense reports for each property at a glance.",
			},
		},
		{
			ID:       uuid.New(),
			Name:     "Maintenance Tracking",
			Title:    "Manage Properties & Maintenance Effortlessly",
			Subtitle: "Maintenance & Repairs",
			Details: []string{
				"Keep a complete record of all your properties in one dashboard.",
				"Organize and assign maintenance & repar tasks to workers.",
				"Let tenants report issues directly from their public tentant portal",
				"Track progress in real time with built-in task management.",
			},
			Description:      "Track, assign, and monitor tasks for contractors - wth tenant-reported issues sent straight to your dashboard.",
			QuickDescription: "Track maintenance & repairs",
		},
		{
			ID:          uuid.New(),
			Name:        "Appointment Management",
			Title:       "Schedule & Track Appointments",
			Subtitle:    "Appointment Scheduling",
			Description: "Organize open houses and walk-throughs with built-in reminders for you and your tenants.",
			Details: []string{
				"Manage <strong>open house events</strong> for prospective tenants.",
				"Coordinate <strong>walk-through inspections</strong> and send reminders automatically",
				"Optimize and Reschedule appointments with <strong>AI powered Smart Scheduling</strong>",
			},
			QuickDescription: "Schedule open houses & walk-throughs",
		},
		{
			ID:          uuid.New(),
			Name:        "Document Generation",
			Title:       "AI-Powered Document Management",
			Subtitle:    "AI Document Assistant",
			Description: "Generate leases, eviction noitces, and more in minutes - no legal jargon required.",
			Details: []string{
				"Instantly generate leases, eviction notices, and other documents with our AI assistant.",
				"Securely store, organize, and share documents with tenants, contractors, and legal teams.",
			},
			QuickDescription: "AI-generate leases, eviction notices, and other documents",
		},
		{
			ID:          uuid.New(),
			Name:        "Web Pages",
			Title:       "Public Web Pages for Your Properties",
			Subtitle:    "Public Web Pages",
			Description: "Get a property listings page for prospects and a tenant portal for tenant requests.",
			Details: []string{
				"A property listing page where potential tenatns can browse and schedule walk-throughs for your available homes.",
				"A tenant support page where renters can report problems, request maintenance, make payments, and access documents.",
			},
			QuickDescription: "Give tenants a portal to report problems",
		},
		{
			ID:          uuid.New(),
			Name:        "Document Storage",
			Title:       "Secure Document Storage & Sharing",
			Subtitle:    "Secure Centralized Document Hub",
			Description: "Store, organize, and share important documents with tenants, contractors, and team members in one secure location",
			Details: []string{
				"Upload and organize <strong>leases, receipts, and notices</strong> in a centralized hub.",
				"Securely <strong>share documents</strong> with tenants, contractors, and property managers.",
				"Enable <strong>e-signatures</strong> for faster agreements and approvals.",
				"Access documents anytime from your <strong>dashboard or tenant portal</strong>.",
			},
			QuickDescription: "Store, share, and sign documents securely",
		},
	}
}
