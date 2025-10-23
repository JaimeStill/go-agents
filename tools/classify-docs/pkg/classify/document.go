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
	"github.com/JaimeStill/go-agents/tools/classify-docs/pkg/retry"
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
		createClassifier(a, cfg.Processing.Retry),
		progressFunc,
	)

	if err != nil {
		return DocumentClassification{}, fmt.Errorf("failed to classify %s: %w", filename, err)
	}

	return result.Final, nil
}

func createClassifier(
	a agent.Agent,
	retryCfg config.RetryConfig,
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

		updated, err := retry.Do(ctx, retryCfg, func(ctx context.Context, attempt int) (DocumentClassification, error) {
			if attempt > 1 {
				fmt.Fprintf(os.Stderr, "  Retry attempt %d for %s page %d...\n", attempt-1, current.File, page.Number())
			}

			response, err := a.Vision(ctx, prompt, []string{encoded})
			if err != nil {
				return current, err
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
		})

		if err != nil {
			return current, fmt.Errorf("vision request failed for page %d: %w", page.Number(), err)
		}

		return updated, nil
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
   - Look for caveats after "//" separators (NOFORN, REL TO, ORCON, etc.)
   - Look for compartments after "//" (SPECAT, X1, X2, etc.)
   - Look for dates in YYYYMMDD format
   - Example: If you see "SECRET" in header and "NOFORN" in footer, return "SECRET//NOFORN"
   - Example: If you see "SECRET//NOFORN//X1", capture the ENTIRE string

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

5. Update the classification field with the COMPLETE highest-level marking found
6. Add each discovered marking component to markings_found (avoid duplicates)
7. BEFORE assigning confidence, critically verify your findings by asking:

   Self-Check Questions:
   - "Is this text ACTUALLY a classification marking or could it be document content/heading?"
   - "Am I seeing markings in the EXPECTED locations (header/footer/margins)?"
   - "Could there be FADED stamps or light markings I overlooked?"
   - "Did I check BOTH header AND footer thoroughly?"
   - "If I marked this UNCLASSIFIED, am I absolutely certain there are no markings?"

   If you answer NO or UNCERTAIN to any question, you MUST lower confidence.

8. Assign confidence level using these criteria:

   HIGH confidence when ALL of these are true:
   - Clear, unambiguous classification markings are visible
   - Markings are consistent across header/footer/margins
   - No conflicting or questionable markings present
   - All expected components are captured (level + caveats if present)

   MEDIUM confidence when ANY of these are true:
   - Markings are partially faded but still readable
   - Some ambiguity in interpreting portions of the marking
   - Missing expected caveats that might be present but unclear
   - Inconsistent formatting but clear classification level

   LOW confidence when ANY of these are true:
   - Markings are heavily faded or barely visible
   - Unclear whether text is a classification marking or document content
   - Conflicting markings present
   - Cannot determine if document is classified at all

9. Update rationale explaining:
   - What complete markings were found and their exact locations
   - Results of your self-check verification questions
   - Why the assigned confidence level is appropriate based on your verification

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
