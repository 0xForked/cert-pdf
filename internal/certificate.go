package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

type PreGenerateCertificate struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	SchoolID        string          `json:"school_id"`
	MaterialID      string          `json:"material_id"`
	Collection      string          `json:"type"`
	CollectionID    string          `json:"type_id,omitempty"`
	ReferenceNumber string          `json:"reference_number"`
	Metadata        *CourseMetadata `json:"metadata"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       *time.Time      `json:"updated_at"`
	ValidFrom       *time.Time      `json:"valid_from"`
	ValidUntil      *time.Time      `json:"valid_until"`
	// relation data
	Material *Material          `json:"material"`
	Talent   *User              `json:"talent"`
	Course   *CourseProgramItem `json:"course,omitempty"`
	Program  *CourseProgramItem `json:"program,omitempty"`
}

type Material struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Signature *string `json:"admin_signature"`
	Owner     *Owner  `json:"owner_certificate"`
	School    *School `json:"school"`
}

type Owner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
}

type School struct {
	ID          string     `json:"_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	Logo        string     `json:"logo"`
	SocialURL   *SocialURL `json:"social_url,omitempty"`
}

type SocialURL struct {
	Website   string `json:"website_url"`
	LinkedIn  string `json:"linkedin_url"`
	Twitter   string `json:"twitter_url"`
	Facebook  string `json:"facebook_url"`
	Instagram string `json:"instagram_url"`
}

type CourseProgramItem struct {
	ID   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

type CourseMetadata struct {
	CourseID     string                 `json:"course_id"`
	GreenSkillID string                 `json:"green_skill_id"`
	Name         string                 `json:"name"`
	Levels       []*CourseLevelMetadata `json:"levels"`
}

type CourseLevelMetadata struct {
	ID              string     `json:"id"`
	ImageURL        string     `json:"image_url"`
	LearningOutcome string     `json:"learning_outcome"`
	Name            string     `json:"name"`
	Proficiency     int        `json:"proficiency"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Avatar    string `json:"avatar"`
	Location  string `json:"location"`
}

func (c PreGenerateCertificate) GeneratePDF(path string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})
	pdf.AddPage()

	if err := pdf.AddTTFFont("kodchasan_regular", "./assets/fonts/kodchasan/Regular.ttf"); err != nil {
		return err
	}
	if err := pdf.AddTTFFont("kodchasan_bold", "./assets/fonts/kodchasan/Bold.ttf"); err != nil {
		return err
	}
	if err := pdf.AddTTFFont("kodchasan_italic", "./assets/fonts/kodchasan/Italic.ttf"); err != nil {
		return err
	}

	setFont := func(style string, fontSize int) error {
		switch style {
		case "bold":
			return pdf.SetFont("kodchasan_bold", "", fontSize)
		case "italic":
			return pdf.SetFont("kodchasan_italic", "", fontSize)
		default:
			return pdf.SetFont("kodchasan_regular", "", fontSize)
		}
	}

	centerText := func(text, style string, y float64, fontSize int, r, g, b uint8) (float64, float64, error) {
		if err := setFont(style, fontSize); err != nil {
			return 0, 0, err
		}
		pdf.SetTextColor(r, g, b)
		textWidth, err := pdf.MeasureTextWidth(text)
		if err != nil {
			return 0, 0, err
		}
		x := (gopdf.PageSizeA4Landscape.W - textWidth) / 2
		pdf.SetXY(x, y)
		if err := pdf.Text(text); err != nil {
			return 0, 0, err
		}
		pdf.SetTextColor(0, 0, 0)
		return x, textWidth, nil
	}

	if _, _, err := centerText(
		"C E R T I F I C A T E   O F   C O M P L E T I O N",
		"regular", 40, 12, 150, 111, 2,
	); err != nil {
		return err
	}

	if _, _, err := centerText(
		"This is to certify that", "regular",
		60, 10, 0, 0, 0,
	); err != nil {
		return err
	}

	yPos := 90.0
	fontSize := 16
	startX, textWidth, err := centerText(c.Talent.FullName, "italic", 90, 16, 0, 0, 0)
	if err != nil {
		return err
	}
	underlineY := yPos + float64(fontSize) - 10
	pdf.SetLineWidth(0.5)
	pdf.Line(startX, underlineY, startX+textWidth, underlineY)

	if _, _, err := centerText(
		"has successfully completed online [non-credit]",
		"regular", 120, 10, 0, 0, 0,
	); err != nil {
		return err
	}

	if _, _, err := centerText(
		strings.ToUpper(c.Metadata.Name), "bold", 155,
		18, 0, 0, 0,
	); err != nil {
		return err
	}

	if _, _, err := centerText(
		"A course delivered by School Name and offered through Skilledin Green platform",
		"regular", 185, 10, 0, 0, 0,
	); err != nil {
		return err
	}

	pdf.SetXY(30, 220)
	if err := pdf.Text("Completed proficiency levels and associated knowledge:"); err != nil {
		return err
	}

	nextY := 225.0
	const (
		col1Width = 40
		col2Width = 60
		col3Width = 100
		maxWords  = 10
	)
	for _, level := range c.Metadata.Levels {
		nextY += 25
		pdf.SetXY(30, nextY)
		updatedAt := level.UpdatedAt.Format("02 January 2006")
		proficiency := fmt.Sprintf("Level %d (%s)", level.Proficiency, capitalizeFirst(level.Name))
		learningOutcome := level.LearningOutcome

		if err := pdf.Text(updatedAt); err != nil {
			return err
		}
		pdf.SetXY(pdf.GetX()+col1Width, nextY)

		if err := pdf.Text(proficiency); err != nil {
			return err
		}
		pdf.SetXY(pdf.GetX()+col2Width, nextY)

		// Handle the learning outcome with splitting into lines of maxWords
		learningOutcomeY := nextY
		words := strings.Fields(learningOutcome) // Split the string into words
		for i := 0; i < len(words); i += maxWords {
			// Get the next part of the learning outcome
			end := i + maxWords
			if end > len(words) {
				end = len(words) // Ensure we don't go out of bounds
			}
			part := strings.Join(words[i:end], " ") // Join the words back into a string
			// Print the learning outcome part
			if err := pdf.Text(part); err != nil {
				return err
			}
			// Move down for the next part
			learningOutcomeY += 10
			pdf.SetXY((col2Width+col3Width)*2, learningOutcomeY)
		}
	}

	return pdf.WritePdf(fmt.Sprintf("%s/%s.pdf", path, c.ReferenceNumber))
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

// download file first
// then load from file
//if err := pdf.Image(c.Material.School.Logo, 200, 50, nil); err != nil {
//	return err
//}

func downloadFile(filepath string, url string) error {
	// check folder exist or not
	// if not create
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
