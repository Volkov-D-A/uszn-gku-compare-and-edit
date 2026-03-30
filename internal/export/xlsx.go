package export

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"uszn-gku-compare-and-edit/internal/domain"
)

type sheet struct {
	Name string
	Rows [][]cell
}

type cell struct {
	Value  string
	Number bool
}

func WriteAnalysisXLSX(path string, report domain.AnalysisReport) error {
	if filepath.Ext(path) == "" {
		path += ".xlsx"
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	sheets := buildSheets(report)

	if err := writeZipFile(writer, "[Content_Types].xml", contentTypesXML(len(sheets))); err != nil {
		return err
	}
	if err := writeZipFile(writer, "_rels/.rels", rootRelsXML()); err != nil {
		return err
	}
	if err := writeZipFile(writer, "docProps/app.xml", appPropsXML(sheets)); err != nil {
		return err
	}
	if err := writeZipFile(writer, "docProps/core.xml", corePropsXML()); err != nil {
		return err
	}
	if err := writeZipFile(writer, "xl/workbook.xml", workbookXML(sheets)); err != nil {
		return err
	}
	if err := writeZipFile(writer, "xl/_rels/workbook.xml.rels", workbookRelsXML(len(sheets))); err != nil {
		return err
	}
	if err := writeZipFile(writer, "xl/styles.xml", stylesXML()); err != nil {
		return err
	}

	for index, sheet := range sheets {
		name := fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)
		if err := writeZipFile(writer, name, worksheetXML(sheet)); err != nil {
			return err
		}
	}

	return writer.Close()
}

func buildSheets(report domain.AnalysisReport) []sheet {
	summary := sheet{
		Name: "Сводка",
		Rows: [][]cell{
			{{Value: "Показатель"}, {Value: "Значение"}},
			{{Value: "Поставщик"}, {Value: report.Meta.ProviderName}},
			{{Value: "Предыдущий месяц"}, {Value: report.Meta.PreviousMonth}},
			{{Value: "Текущий месяц"}, {Value: report.Meta.CurrentMonth}},
			{{Value: "Записей в предыдущем месяце"}, {Value: strconv.Itoa(report.Meta.PreviousRecords), Number: true}},
			{{Value: "Записей в текущем месяце"}, {Value: strconv.Itoa(report.Meta.CurrentRecords), Number: true}},
			{{Value: "Изменилось тарифов"}, {Value: strconv.Itoa(report.Summary.TariffChanges), Number: true}},
			{{Value: "Появилось услуг"}, {Value: strconv.Itoa(report.Summary.AppearedServices), Number: true}},
			{{Value: "Исчезло услуг"}, {Value: strconv.Itoa(report.Summary.DisappearedServices), Number: true}},
			{{Value: "Появилось домов"}, {Value: strconv.Itoa(report.Summary.AppearedHouses), Number: true}},
			{{Value: "Исчезло домов"}, {Value: strconv.Itoa(report.Summary.DisappearedHouses), Number: true}},
			{{Value: "Аномалий"}, {Value: strconv.Itoa(report.Summary.Anomalies), Number: true}},
		},
	}

	tariffs := sheet{
		Name: "Тарифы",
		Rows: [][]cell{{{Value: "VID_USL"}, {Value: "NAME_USL"}, {Value: "Тариф был"}, {Value: "Тариф стал"}}},
	}
	for _, item := range report.TariffChanges {
		tariffs.Rows = append(tariffs.Rows, []cell{
			{Value: item.VidUsl},
			{Value: item.NameUsl},
			{Value: formatFloat(item.PreviousTariff), Number: true},
			{Value: formatFloat(item.CurrentTariff), Number: true},
		})
	}

	services := sheet{
		Name: "Услуги",
		Rows: [][]cell{{{Value: "Тип"}, {Value: "VID_USL"}, {Value: "NAME_USL"}}},
	}
	for _, item := range report.ServiceChanges {
		services.Rows = append(services.Rows, []cell{
			{Value: item.Type},
			{Value: item.VidUsl},
			{Value: item.NameUsl},
		})
	}

	houses := sheet{
		Name: "Дома",
		Rows: [][]cell{{{Value: "Тип"}, {Value: "Адрес"}}},
	}
	for _, item := range report.HouseChanges {
		houses.Rows = append(houses.Rows, []cell{
			{Value: item.Type},
			{Value: item.Address},
		})
	}

	anomalies := sheet{
		Name: "Аномалии",
		Rows: [][]cell{{{Value: "VID_USL"}, {Value: "NAME_USL"}, {Value: "Предыдущая сумма"}, {Value: "Текущая сумма"}, {Value: "Изменение"}, {Value: "Изменение %"}, {Value: "Порог %"}}},
	}
	for _, item := range report.Anomalies {
		deltaPercent := ""
		if item.DeltaPercent != nil {
			deltaPercent = formatFloat(*item.DeltaPercent)
		}
		anomalies.Rows = append(anomalies.Rows, []cell{
			{Value: item.VidUsl},
			{Value: item.NameUsl},
			{Value: formatFloat(item.PreviousAmount), Number: true},
			{Value: formatFloat(item.CurrentAmount), Number: true},
			{Value: formatFloat(item.DeltaAmount), Number: true},
			{Value: deltaPercent, Number: deltaPercent != ""},
			{Value: formatFloat(item.ThresholdPercent), Number: true},
		})
	}

	return []sheet{summary, tariffs, services, houses, anomalies}
}

func writeZipFile(writer *zip.Writer, name string, body string) error {
	entry, err := writer.Create(name)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", name, err)
	}
	if _, err := entry.Write([]byte(body)); err != nil {
		return fmt.Errorf("write zip entry %s: %w", name, err)
	}
	return nil
}

func contentTypesXML(sheetCount int) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	builder.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	builder.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	builder.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	builder.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	builder.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	builder.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	for i := 1; i <= sheetCount; i++ {
		builder.WriteString(fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i))
	}
	builder.WriteString(`</Types>`)
	return builder.String()
}

func rootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
}

func appPropsXML(sheets []sheet) string {
	titles := make([]string, 0, len(sheets))
	for _, sheet := range sheets {
		titles = append(titles, xmlEscape(sheet.Name))
	}

	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">`)
	builder.WriteString(`<Application>Codex</Application><DocSecurity>0</DocSecurity><ScaleCrop>false</ScaleCrop>`)
	builder.WriteString(`<HeadingPairs><vt:vector size="2" baseType="variant"><vt:variant><vt:lpstr>Worksheets</vt:lpstr></vt:variant><vt:variant><vt:i4>`)
	builder.WriteString(strconv.Itoa(len(sheets)))
	builder.WriteString(`</vt:i4></vt:variant></vt:vector></HeadingPairs>`)
	builder.WriteString(`<TitlesOfParts><vt:vector size="`)
	builder.WriteString(strconv.Itoa(len(sheets)))
	builder.WriteString(`" baseType="lpstr">`)
	for _, title := range titles {
		builder.WriteString(`<vt:lpstr>`)
		builder.WriteString(title)
		builder.WriteString(`</vt:lpstr>`)
	}
	builder.WriteString(`</vt:vector></TitlesOfParts></Properties>`)
	return builder.String()
}

func corePropsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:creator>Codex</dc:creator>
  <cp:lastModifiedBy>Codex</cp:lastModifiedBy>
</cp:coreProperties>`
}

func workbookXML(sheets []sheet) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for index, sheet := range sheets {
		builder.WriteString(fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEscape(sheet.Name), index+1, index+1))
	}
	builder.WriteString(`</sheets></workbook>`)
	return builder.String()
}

func workbookRelsXML(sheetCount int) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= sheetCount; i++ {
		builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i, i))
	}
	builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheetCount+1))
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func stylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts>
  <fills count="1"><fill><patternFill patternType="none"/></fill></fills>
  <borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
}

func worksheetXML(sheet sheet) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for rowIndex, row := range sheet.Rows {
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowIndex+1))
		for colIndex, c := range row {
			ref := cellRef(colIndex+1, rowIndex+1)
			builder.WriteString(cellXML(ref, c))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func cellXML(ref string, c cell) string {
	if c.Number {
		return fmt.Sprintf(`<c r="%s"><v>%s</v></c>`, ref, xmlEscape(c.Value))
	}

	var buffer bytes.Buffer
	encoder := xml.NewEncoder(&buffer)
	_ = encoder.EncodeToken(xml.CharData([]byte(c.Value)))
	_ = encoder.Flush()
	return fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, ref, strings.TrimSpace(buffer.String()))
}

func cellRef(col int, row int) string {
	return columnName(col) + strconv.Itoa(row)
}

func columnName(col int) string {
	result := ""
	for col > 0 {
		col--
		result = string(rune('A'+(col%26))) + result
		col /= 26
	}
	return result
}

func xmlEscape(value string) string {
	var buffer bytes.Buffer
	encoder := xml.NewEncoder(&buffer)
	_ = encoder.EncodeToken(xml.CharData([]byte(value)))
	_ = encoder.Flush()
	return strings.TrimSpace(buffer.String())
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
