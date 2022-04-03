# rea
document renderer using OpenDocument templates

## Notes
### Zipping OpenDocuments
```bash
zip -0 Basic2.ott mimetype
zip -0 -r -u Basic2.ott *
unzip -lv Basic2.ott
libreoffice Basic2.ott
```

### CLI Design
```plaintext
# Simple human CLI usage
rea template -t template.ott -i data.yaml -o document.odt
rea render -t template.ott -i data.yaml -o document.pdf

# Advanced usage
rea render -f job.yaml
```

Shorthands:
- `-d`: Debug bundle file?
- `-b`: Bundle file?

## Bundle file
Contents:
- Job file
- LuaProg
- Input Document
- Input Data
- XML Tree
- Version

## Data file
```
apiVersion: v1alpha1
kind: RenderJob
metadata:
  name: render-job-deadbeef
spec:
  templateFile: template.ott
```
