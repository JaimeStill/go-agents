package classify

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/JaimeStill/go-agents/pkg/agent"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/config"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/document"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/encoding"
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/processing"
)

func ClassifyDocument(
	ctx context.Context,
	cfg config.ClassifyConfig,
	a agent.Agent,
	pdfPath string,
) (DocumentClassification, error) {
	doc, err := document.OpenPDF(pdfPath)
	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to open PDF: %w", err)
	}

	pages, err := doc.ExtractAllPages()
	doc.Close()

	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to extract pages: %w", err)
	}

	filename := filepath.Base(pdfPath)

	initial := DocumentClassification{
		File:                    filename,
		Classification:          "",
		Confidence:              "",
		MarkingsFound:           []string{},
		ClassificationRationale: "",
	}

	progressFunc := func(completed, total int, current DocumentClassification) {
		fmt.Fprintf(os.Stderr, "  Page %d/%d processed\n", completed, total)
	}

	result, err := processing.ProcessWithContext(
		ctx,
		cfg.Processing.Sequential,
		pages,
		initial,
		createClassifier(a),
		progressFunc,
	)

	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to classify %s: %w", filename, err)
	}

	return result.Final, nil
}

func createClassifier(
	a agent.Agent,
) processing.ContextProcessor[DocumentClassification] {
	return func(
		ctx context.Context,
		page document.Page,
		current DocumentClassification,
	) (DocumentClassification, error) {
		if err := ctx.Err(); err != nil {
			return current, fmt.Errorf("context cancelled: %w", err)
		}

		data, err := page.ToImage(document.DefaultImageOptions())
		if err != nil {
			return current, fmt.Errorf("failed to render page %d: %w", page.Number(), err)
		}

		encoded, err := encoding.EncodeImageDataURI(data, document.PNG)
		if err != nil {
			return current, fmt.Errorf("failed to encode page %d: %w", page.Number(), err)
		}

		prompt := buildClassificationPrompt(current)

		response, err := a.Vision(ctx, prompt, []string{encoded})
		if err != nil {
			return current, fmt.Errorf("vision request failed for page %d: %w", page.Number(), err)
		}

		if len(response.Choices) == 0 {
			return current, fmt.Errorf("empty response for page %d", page.Number())
		}

		content := response.Content()
		if strings.TrimSpace(content) == "" {
			return current, fmt.Errorf("received empty classification for page %d", page.Number())
		}

		classification, err := parseClassificationResponse(content)
		if err != nil {
			return current, fmt.Errorf("failed to parse page %d response: %w", page.Number(), err)
		}

		return classification, nil
	}
}

const classificationPromptTemplate = `Current document classification state:

{
	"file": {{.File | printf "%q"}},
	"classification": {{.Classification | printf "%q"}},
	"confidence": {{.Confidence | printf "%q"}},
	"markings_found": [{{range $i, $m := .MarkingsFound}}{{if $i}}, {{end}}{{$m | printf "%q"}}{{end}}],
	"classification_rationale": {{.ClassificationRationale | printf "%q"}}
}
{{if not .Classification}}

This is the first page - initialize the classification.
{{end}}

Analyze this page image and identify classification markings with COMPLETE ACCURACY:

CRITICAL RULES:
1. NEVER abbreviate classification levels - always use full spelling:
   - Write "UNCLASSIFIED" not "U"
   - Write "CONFIDENTIAL" not "C"
   - Write "SECRET" not "S"
   - Write "TOP SECRET" not "TS"

2. ALWAYS capture the COMPLETE marking string including ALL caveats:
   - CRITICAL: Classification components are often SPATIALLY SEPARATED across the page
   - Classification level may appear in one location while caveats appear elsewhere
   - You MUST combine ALL components found anywhere on the page into one complete marking
   - Look for caveats after "//" separators (NOFORN, REL TO, ORCON, etc.)
   - Look for compartments after "//" (SPECAT, X1, X2, etc.)
   - Look for dates in YYYYMMDD format
   - Example: "SECRET" in header + "NOFORN" stamp elsewhere → return "SECRET//NOFORN"
   - Example: If you see "SECRET//NOFORN//X1", capture the ENTIRE string

   SPECIAL EMPHASIS FOR FADED CAVEATS:
   - NOFORN and ORCON stamps are FREQUENTLY faded or barely visible
   - Low contrast does NOT mean invalid - these are legitimate classification components
   - Apply maximum visual scrutiny to detect faint stamps anywhere on the page
   - Examine ALL areas of the document carefully, not just prominent header/footer banners

3. ONLY identify actual classification markings - NOT document content:
   - Banner markings at top/bottom of page (usually centered or in margins)
   - Portion markings in parentheses like (S), (C), (U)
   - Classification labels: "CLASSIFICATION: SECRET", "CAVEATS: NOFORN"
   - Header/footer stamps (even if faded or light)
   - DO NOT classify document titles, section headings, or body text
   - Classification markings are SEPARATE from document content
   - Example: "SEQUENCE OF EVENTS" is a document heading, NOT a classification
   - Example: If you see what looks like "TS" in a heading, verify it's a marking, not part of a word

4. Search the ENTIRE page INCLUDING faint/faded markings:
   - Check headers AND footers carefully
   - Check margins on all sides
   - Look for FADED or LIGHT stamps (common for NOFORN, ORCON, etc.)
   - Look for partially visible markings through redactions
   - Adjust for low-contrast markings - they are still valid
   - Build the complete classification from ALL components found (even faint ones)
   - CRITICAL: If you see faint/faded text but cannot read it clearly, you MUST:
     * Note the presence of unreadable faded marking in your rationale
     * Assign MEDIUM or LOW confidence (never HIGH)
     * State uncertainty about potential caveats in your rationale

5. Update the classification field with the COMPLETE highest-level marking found
6. Add each discovered marking component to markings_found (avoid duplicates)
7. BEFORE assigning confidence, critically verify your findings by asking:

   Self-Check Questions:
   - "Is this text ACTUALLY a classification marking or could it be document content/heading?"
   - "Am I seeing markings in the EXPECTED locations (header/footer/margins)?"
   - "Did I check for SEPARATED components (classification in one place, caveats elsewhere)?"
   - "Could there be FADED caveat stamps (NOFORN, ORCON) that I cannot detect at this resolution?"
   - "Did I examine ALL areas of the page thoroughly, including low-contrast regions?"
   - "If I marked this UNCLASSIFIED, am I absolutely certain there are no markings?"
   - "CRITICAL: If this document is SECRET/CONFIDENTIAL/TOP SECRET but I see NO caveats, could there be faded caveats I'm missing?"

   If you answer NO or UNCERTAIN to any question, you MUST assign MEDIUM or LOW confidence.

8. Assign confidence level conservatively - it is BETTER to be uncertain than overconfident:

   HIGH confidence ONLY when ALL of these are true:
   - Classification markings are crystal clear with NO fading whatsoever
   - ALL components are easily readable with perfect clarity
   - No ambiguity exists about ANY part of the marking
   - For classified documents: EITHER clear caveats are visible OR you have explicit proof no caveats exist
   - You are absolutely certain no additional caveats could be present
   - IMPORTANT: If you detected ANY faintness, you CANNOT assign HIGH confidence
   - IMPORTANT: If document is SECRET/CONFIDENTIAL/TOP SECRET with no visible caveats, you CANNOT assign HIGH confidence

   MEDIUM confidence when ANY of these are true (THIS IS THE EXPECTED DEFAULT):
   - You detected any faintness, fading, or low-contrast markings
   - Document shows SECRET/CONFIDENTIAL/TOP SECRET but NO caveats visible (caveats are common and may be faded)
   - You see "SECRET" clearly but cannot rule out additional faded caveats
   - Markings are readable but not perfectly clear
   - You combined components from different locations
   - Any self-check question raised doubt or uncertainty
   - You searched for caveats but the quality makes you uncertain

   IMPORTANT: Classified documents (SECRET, CONFIDENTIAL, TOP SECRET) often have caveats.
   If you see ONLY the classification level with no caveats, assume MEDIUM confidence because
   caveats may be present but too faded for you to detect.

   LOW confidence when ANY of these are true:
   - Markings are heavily faded or barely visible
   - You cannot confidently read key components
   - Conflicting markings present
   - Unclear whether text is classification marking or document content

   REMEMBER: MEDIUM confidence is valuable and acceptable. It flags documents for human review.
   Being conservative protects against misclassification.

9. Update rationale explaining:
   - What complete markings were found and their exact locations
   - Results of your self-check verification questions
   - Why the assigned confidence level is appropriate based on your verification
   - If assigning HIGH confidence, explicitly confirm NO faintness was detected
   - If you saw faint markings, explain why MEDIUM/LOW confidence was assigned

Return ONLY the updated DocumentClassification as valid JSON.
Do NOT wrap in markdown code fences.
Do NOT add conversational text - just the JSON object.`

var promptTemplate *template.Template

func init() {
	promptTemplate = template.Must(template.New("classification").Parse(classificationPromptTemplate))
}

func buildClassificationPrompt(current DocumentClassification) string {
	var buf bytes.Buffer
	if err := promptTemplate.Execute(&buf, current); err != nil {
		return "Analyze this page and provide classification as JSON."
	}

	return buf.String()
}
