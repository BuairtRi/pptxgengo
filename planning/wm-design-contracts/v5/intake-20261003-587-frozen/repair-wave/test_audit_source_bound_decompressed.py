"""Adversarial checks for the narrowly scoped XLSX audit normalization."""
import importlib.util
import io
import unittest
import warnings
import zipfile
from pathlib import Path

SPEC = importlib.util.spec_from_file_location('audit', Path(__file__).with_name('audit-source-bound-decompressed.py'))
AUDIT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(AUDIT)
CORE = (b'<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" '
        b'xmlns:dcterms="http://purl.org/dc/terms/"><cp:creator>Author</cp:creator>'
        b'<dcterms:created>2026-10-03T22:00:00Z</dcterms:created>'
        b'<dcterms:modified>2026-10-03T22:00:00Z</dcterms:modified></cp:coreProperties>')


def package(core=CORE, sheet=b'<worksheet>1.25</worksheet>', style=b'<style>IBM Plex Sans</style>', year=2000):
    output = io.BytesIO()
    with zipfile.ZipFile(output, 'w') as archive:
        for name, data in [('docProps/core.xml', core), ('xl/worksheets/sheet1.xml', sheet), ('xl/styles.xml', style)]:
            member = zipfile.ZipInfo(name, (year, 1, 1, 0, 0, 0))
            archive.writestr(member, data)
    return output.getvalue()


class NarrowNormalization(unittest.TestCase):
    def test_timestamp_and_zip_framing_only(self):
        a = package()
        b = package(CORE.replace(b'22:00:00', b'22:00:01'), year=2020)
        self.assertNotEqual(a, b)
        self.assertEqual(AUDIT.xlsx_members(a, True), AUDIT.xlsx_members(b, True))
        self.assertNotEqual(AUDIT.xlsx_members(a), AUDIT.xlsx_members(b))

    def test_copy_data_font_and_other_metadata_are_not_normalized(self):
        original = AUDIT.xlsx_members(package(), True)
        for changed in [package(CORE.replace(b'Author', b'Other')), package(sheet=b'<worksheet>1.26</worksheet>'),
                        package(style=b'<style>Other Font</style>')]:
            self.assertNotEqual(original, AUDIT.xlsx_members(changed, True))

    def test_missing_or_duplicate_timestamp_rejected(self):
        for changed in [CORE.replace(b'<dcterms:created>2026-10-03T22:00:00Z</dcterms:created>', b''),
                        CORE.replace(b'</cp:coreProperties>', b'<dcterms:created>x</dcterms:created></cp:coreProperties>')]:
            with self.assertRaises(ValueError):
                AUDIT.core_metadata(changed)

    def test_duplicate_zip_member_rejected(self):
        output = io.BytesIO()
        with warnings.catch_warnings():
            warnings.simplefilter('ignore', UserWarning)
            with zipfile.ZipFile(output, 'w') as archive:
                archive.writestr('xl/worksheet.xml', b'one')
                archive.writestr('xl/worksheet.xml', b'two')
        with self.assertRaises(ValueError):
            AUDIT.xlsx_members(output.getvalue(), True)


if __name__ == '__main__':
    unittest.main()
