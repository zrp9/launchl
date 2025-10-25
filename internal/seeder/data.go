// Package seeder initial data for db
package seeder

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/zrp9/launchl/internal/domain/core"
)

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

type AppSurvey struct {
	ID          uuid.UUID
	Questions   []AppQuestion
	Intro       string
	ClosingText string
	Version     string
	Name        string
	Active      bool
}

type AppQuestion struct {
	ID           uuid.UUID
	SurveyID     uuid.UUID
	QuestionType string
	Options      []AppQuestionOption
	Prompt       string
	Position     int
	Active       bool
	Required     bool
	MetaData     json.RawMessage
}

type AppQuestionOption struct {
	ID         uuid.UUID
	QuestionID uuid.UUID
	Position   int
	Label      string
	Value      string
}

type Permission uint8

const (
	Read       Permission = 1 << iota // 0001 binary
	Write                             // 0010
	Execute                           // 0100
	Admin                             // 1000
	SuperAdmin                        // 0011
)

func GetPermission(p Permission) string {
	permissions := make([]string, 0)
	if p&Read != 0 {
		permissions = append(permissions, "read")
	}

	if p&Write != 0 {
		permissions = append(permissions, "write")
	}

	if p&Execute != 0 {
		permissions = append(permissions, "execute")
	}

	if p&Admin != 0 {
		permissions = append(permissions, "admin")
	}

	if p&SuperAdmin != 0 {
		permissions = append(permissions, "super-admin")
	}

	return strings.Join(permissions, ",")
}

func RemovePermission(curPermissions Permission, perm Permission) Permission {
	return curPermissions &^ perm
}

func AddPermission(curPermissions Permission, perm Permission) Permission {
	return curPermissions | perm
}

func GrantPermissions(perms ...Permission) string {
	var permissions Permission
	for _, p := range perms {
		permissions |= p
	}

	return GetPermission(permissions)
}

//  maybe make a lambda that can be called with strings to save to s3 for feature

func GetAppFeatures() []AppFeature {
	return []AppFeature{
		{
			ID:               uuid.New(),
			Name:             "AI-Powered Bid Generator",
			Title:            "Generate Accurate Bids Fast",
			Subtitle:         "Describe the job; get a complete estimate",
			QuickDescription: "Create detailed, accurate job estimates in seconds - just describe the project and let AI handle the rest",
			Description:      "Say goodbye to guesswork and forgotten materials. Our AI-driven Bid Generator takes a simple description like 'vinyl siding for a 2-story home' and automatically builds a full, itemized estimate - down to the nails, screws, and caulk. It factors in current material prices from major retailers (like Lowe's) and lets you adjust waste factors, labor rates, and margins before exporting a professional-grade PDF bid ready to send or convert into a work order.",
			Details: []string{
				"Use AI to instantly generate <strong>material lists</strong> and <strong>labor estimates</strong> from plain-language descriptions.",
				"Automatically include <strong>small consumables</strong> like nails, screws, and caulk.",
				"Adjust <strong>waste factors</strong>, <strong>labor rates</strong>, and <strong>margins</strong> to fine-tune totals.",
				"Export your finished bid as a <strong>PDF or job order</strong> with one click.",
			},
		},
		{
			ID:       uuid.New(),
			Name:     "Smart Expense & Tax Tracker",
			Title:    "Property Expenses & Taxes",
			Subtitle: "Track spend by property, export at tax time",
			Details: []string{
				"Categorize and track <strong>expenses</strong> across all properties.",
				"Upload and store <strong>receipts</strong> and <strong>invoices</strong> digitally.",
				"View <strong>income vs. expenses</strong> by property and by year.",
				"Generate <strong>tax-ready reports</strong> to simplify filing season.",
			},
			QuickDescription: "Track every dollar spend on your properties and generate clean, ready-to-file tax summaries.",
			Description:      "From maintenance costs to material receipts, the Smart Expense Tracker keeps your books organized property-by-property. Easily categorize expenses (repairs, improvements, utilities, insurance, taxes) and attach receipts with one click. At tax time, export your data irectly into IRS friendly formats - including Schedule E reports - or share with your accountant. The system auto-totals deductible amounts, helping you save time and money while staying compliant.",
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
			Name:        "Automated Tax Report Export",
			Title:       "One-Click Tax Exports",
			Subtitle:    "Schedule E, CSV, and PDFs",
			Description: "No more scrambling through spreadsheets or receipts every April. Our tax export tool compiles your income and expense data across all properties into a clean report you can hand directly to your CPA. It automatically separates deductible categories, flags potential write-offs, and provides downloadable PDFs and CSVs tailored for Schedule E or other small-business filing formats.",
			Details: []string{
				"Automatically summarize <strong>income and expenses</strong> by category.",
				"Export reports in <strong>Schedule E</strong> or <strong>CSV/PDF</strong> formats.",
				"Highlight <strong>deductible expenses</strong> and potential write-offs.",
				"Save time and ensure <strong>tax compliance</strong> with every export.",
			},
			QuickDescription: "Turn a year’s worth of property data into ready-to-file tax documents with one click.",
		},
		{
			ID:       uuid.New(),
			Name:     "Property Listing Portal",
			Title:    "Public Listings Page",
			Subtitle: "Show vacancies, collect interest",
			Details: []string{
				"Create <strong>public listings</strong> for available units.",
				"Add <strong>photos, rent details, and availability dates</strong> easily.",
				"Allow tenants to <strong>apply or request walk-throughs</strong> directly.",
				"Listings update automatically as units are leased.",
			},
			Description:      "Showcase your available units and attract qualified tenants — all in one place.",
			QuickDescription: "Each landlord gets a personalized, mobile-friendly listings page to display current and upcoming vacancies. Add property photos, rent details, and availability dates. Tenants can submit applications, request walk-throughs, or join a waitlist directly from your public portal. Listings update automatically as units are leased, keeping your online presence fresh without extra work.",
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
		{
			ID:          uuid.New(),
			Name:        "Tenant Payments & Ledger",
			Title:       "Online Rent & Ledger",
			Subtitle:    "Collect payments without the paperwork",
			Description: "Tenants can pay securely through integrated online payment options (credit, debit, or ACH), while landlords get instant updates to their income ledger. Payments are automatically tied to the right property and tenant account. Generate monthly summaries, flag late payments, and sync everything with your tax tracker for effortless bookkeeping.",
			Details: []string{
				"Accept <strong>credit, debit, or ACH</strong> payments securely.",
				"Automatically update your <strong>income ledger</strong> per tenant.",
				"Generate <strong>monthly summaries</strong> and flag late payments.",
				"Sync payment data with your <strong>expense and tax reports</strong>.",
			},
			QuickDescription: "Collect rent and track payments the modern way — fast, transparent, and paper-free.",
		},
		{
			ID:          uuid.New(),
			Name:        "Tenant Task & Maintenance Reporting",
			Title:       "Maintenance Tickets",
			Subtitle:    "Report, track, and resolve issues",
			Description: "Tenants can pay securely through integrated online payment options (credit, debit, or ACH), while landlords get instant updates to their income ledger. Payments are automatically tied to the right property and tenant account. Generate monthly summaries, flag late payments, and sync everything with your tax tracker for effortless bookkeeping.Whether it’s a leaky faucet or broken window, tenants can open a maintenance ticket right from their portal. Each task includes photos, notes, and status updates. Landlords can assign jobs to contractors or internal crews, record costs, and link completed work directly to property expenses — giving a complete picture of upkeep and profitability.",
			Details: []string{
				"Allow tenants to <strong>report maintenance issues</strong> with notes and photos.",
				"Track issue status from <strong>reported to resolved</strong>.",
				"Assign jobs to <strong>contractors or in-house staff</strong> easily.",
				"Link completed work to <strong>property expenses</strong> automatically.",
			},
			QuickDescription: "Tenants can submit issues, upload photos, and track repair progress in real time.",
		},
		{
			ID:          uuid.New(),
			Name:        "Services you offer",
			Title:       "Restoration & Remodeling Services",
			Subtitle:    "Showcase the jobs you take on",
			Description: "Create a professional services page highlighting your company’s expertise in restoration and remodeling. List the types of projects you handle — such as siding, bathroom tiling, porch builds, and custom carpentry — complete with photo galleries, estimated pricing ranges, and direct inquiry links. It doubles as both a marketing portfolio and a quick-quote launcher inside your Bid Generator.",
			Details: []string{
				"Highlight <strong>services</strong> like siding, tiling, and remodels.",
				"Show <strong>photos, pricing ranges, and descriptions</strong> of your work.",
				"Let clients <strong>inquire directly</strong> or request quotes.",
				"Integrate with the <strong>Bid Generator</strong> for instant estimates.",
			},
			QuickDescription: "Show potential clients exactly what kind of work you do — from siding to bathroom remodels.",
		},
	}
}

func GetAppSurvey() core.Survey {
	surveyID := uuid.New()
	return core.Survey{
		ID:          surveyID,
		Questions:   getQuestions(surveyID),
		Version:     "1.0.0",
		Name:        "Initial Validation Survey",
		Intro:       "Thanks for your interest in Lessor - a simple property and project management app for small landlords and restoration pros. This quick survey (under 2 minutes) helps us shape the product around what matters most to you.",
		ClosingText: "📣 Thanks for helping shape Lessor! Your feedback directly influences what we build. Want to move up the early-access list? Share your referral link on the next page.",
		Active:      true,
	}
}

func GetAppRoles() []AppRoles {
	return []AppRoles{
		{
			ID:          uuid.New(),
			Name:        "subscriber",
			Permissions: "read-write",
		},
		{
			ID:          uuid.New(),
			Name:        "guest",
			Permissions: "read",
		},
		{
			ID:          uuid.New(),
			Name:        "SuperAdmin",
			Permissions: "all",
		},
	}
}

func getQuestions(surveyID uuid.UUID) []core.Question {
	questionIDs := make([]uuid.UUID, 9)

	for _ = range 8 {
		questionIDs = append(questionIDs, uuid.New())
	}

	return []core.Question{
		{
			ID:           questionIDs[0],
			Position:     0,
			SurveyID:     surveyID,
			QuestionType: "check",
			Prompt:       "What best describes you?",
			Active:       true,
			Required:     true,
			// use meta data to determine the other field to open a texxt field
			MetaData: json.RawMessage(`{"hasOtherField": true, "otherField": 5 }`),
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[0],
					Position:   0,
					Label:      "Landlord / property owner",
					Value:      "landlord / property owner",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[0],
					Position:   1,
					Label:      "Property manager",
					Value:      "property manager",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[0],
					Position:   2,
					Label:      "Landlord / contractor",
					Value:      "landlord / contractor",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[0],
					Position:   3,
					Label:      "Contractor / restoration bussiness owner",
					Value:      "contractor / restoration business owner",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[0],
					Position:   4,
					Label:      "Tenant",
					Value:      "tenant",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[0],
					Position:   5,
					Label:      "Other (please specify)",
					// value in text box shown will append to value
					Value: "other:",
				},
			},
		},
		{
			ID:           questionIDs[1],
			Position:     1,
			SurveyID:     surveyID,
			QuestionType: "check",
			Prompt:       "How many rental properties do you currently manage?",
			Active:       true,
			Required:     true,
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[1],
					Position:   0,
					Label:      "1",
					Value:      "1",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[1],
					Position:   1,
					Label:      "2-5",
					Value:      "2-5",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[1],
					Position:   2,
					Label:      "6-10",
					Value:      "6-10",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[1],
					Position:   3,
					Label:      "11-20",
					Value:      "11-20",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[1],
					Position:   4,
					Label:      "21+",
					Value:      "21+",
				},
			},
		},
		{
			ID:           questionIDs[2],
			Position:     2,
			SurveyID:     surveyID,
			QuestionType: "multi-check",
			Prompt:       "What tools do you currently use to track property expenses, jobs, or taxes?",
			Active:       true,
			Required:     true,
			MetaData:     json.RawMessage(`{"hasOtherField": true, "otherField": 5, "maxSelect": 3}`),
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[2],
					Position:   0,
					Label:      "Spreadsheets",
					Value:      "spreadsheets",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[2],
					Position:   1,
					Label:      "QuickBooks",
					Value:      "quickBooks",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[2],
					Position:   2,
					Label:      "Buildium",
					Value:      "buildium",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[2],
					Position:   3,
					Label:      "Paper notebooks",
					Value:      "paperNotebooks",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[2],
					Position:   4,
					Label:      "Nothing",
					Value:      "nothing",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[2],
					Position:   5,
					Label:      "Other",
					Value:      "other:",
				},
			},
		},
		{
			ID:           questionIDs[3],
			Position:     3,
			SurveyID:     surveyID,
			QuestionType: "multi-check",
			Prompt:       "What are your biggest frustrations when managing properties or jobs?",
			Required:     true,
			Active:       true,
			MetaData:     json.RawMessage(`{"hasOtherField": true, "otherField": 6, "maxSelect": 3}`),
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   0,
					Label:      "Keeping up with maintenance and repairs",
					Value:      "keeping up with maintenance & repairs",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   1,
					Label:      "Tracking expenses and receipts",
					Value:      "tracking expenss & receipts",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   2,
					Label:      "Estimating project costs accurately",
					Value:      "estimating project costs",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   3,
					Label:      "Filing taxes / preparing Schedule E",
					Value:      "filing taxes / preparing schedule E",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   4,
					Label:      "Collecting rent and payments",
					Value:      "collecting rent & payments",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   5,
					Label:      "Communicating with tenants or contractors",
					Value:      "communicating with tenants & contractors",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[3],
					Position:   6,
					Label:      "Something else",
					Value:      "something else:",
				},
			},
		},
		{
			ID:           questionIDs[4],
			SurveyID:     surveyID,
			Position:     4,
			QuestionType: "multi-check",
			Prompt:       "Which of these features would you find most useful?",
			Required:     true,
			Active:       true,
			MetaData:     json.RawMessage(`{"hasOtherField": true, "otherField": 7, "maxSelect": 3}`),
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   0,
					Label:      "AI-powered bid and estimate generator",
					Value:      "ai bid & estimate generator",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   1,
					Label:      "Expense and tax tracking per property",
					Value:      "property expense & tax tracking",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   2,
					Label:      "Tenant payment portal",
					Value:      "tenant payment portal",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   3,
					Label:      "Maintenance request tracking",
					Value:      "maintenance request tracking",
				},
				{
					ID: uuid.New(),

					QuestionID: questionIDs[4],
					Position:   4,
					Label:      "Property listing and walkthrough scheduling",
					Value:      "property listing & walkthrough scheduling",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   5,
					Label:      "Tax export reports (Schedule E, CSV, PDF)",
					Value:      "tax export reports: schedule e, csv, pdf",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   6,
					Label:      "Contractor/job management board",
					Value:      "contractor / job management board",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[4],
					Position:   7,
					Label:      "Other",
					Value:      "other:",
				},
			},
		},
		{
			ID:           questionIDs[5],
			SurveyID:     surveyID,
			Position:     5,
			QuestionType: "check",
			Prompt:       "How likely are you to try a tool like Lessor in the next 6 months?",
			Required:     true,
			Active:       true,
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[5],
					Position:   0,
					Label:      "Very likely",
					Value:      "very likely",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[5],
					Position:   1,
					Label:      "Somewhat likely",
					Value:      "somewhat likely",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[5],
					Position:   2,
					Label:      "Not sure yet",
					Value:      "not sure yet",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[5],
					Position:   3,
					Label:      "Unlikely",
					Value:      "unlikely",
				},
			},
		},
		{
			ID:           questionIDs[6],
			SurveyID:     surveyID,
			QuestionType: "text",
			Position:     6,
			Prompt:       "What would make you more likely to try Lessor?",
			Required:     true,
			Active:       true,
		},
		{
			ID:           questionIDs[7],
			SurveyID:     surveyID,
			QuestionType: "text",
			Position:     7,
			Prompt:       "If you could wave a magic want and add one feature to Lessor what would it be?",
			Required:     true,
			Active:       true,
		},
		{
			ID:           questionIDs[8],
			SurveyID:     surveyID,
			QuestionType: "check",
			Position:     8,
			Prompt:       "Would you like early access when Lessor launches?",
			Required:     true,
			Active:       true,
			Options: []core.SurveyQuestionOption{
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[8],
					Position:   0,
					Label:      "Yes - keep me on the early access list",
					Value:      "yes keep me on list",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[8],
					Position:   1,
					Label:      "Maybe later",
					Value:      "maybe later",
				},
				{
					ID:         uuid.New(),
					QuestionID: questionIDs[8],
					Position:   2,
					Label:      "No thanks",
					Value:      "no thanks",
				},
			},
		},
	}
}
