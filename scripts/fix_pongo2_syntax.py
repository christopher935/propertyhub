#!/usr/bin/env python3
import os
import re
import sys

def convert_go_to_pongo2(content):
    """Convert Go html/template syntax to Pongo2 syntax"""
    
    # Remove {{define "name"}} and {{end}} at top/bottom (these are Go block definitions)
    content = re.sub(r'\{\{define\s+"[^"]+"\}\}\s*', '', content)
    content = re.sub(r'\s*\{\{end\}\}\s*$', '', content)
    
    # Convert {{.Variable}} to {{ Variable }}
    content = re.sub(r'\{\{\s*\.(\w+)\s*\}\}', r'{{ \1 }}', content)
    
    # Convert {{.Object.Field}} to {{ Object.Field }}
    content = re.sub(r'\{\{\s*\.(\w+)\.(\w+)\s*\}\}', r'{{ \1.\2 }}', content)
    
    # Convert {{.Object.Field.SubField}} to {{ Object.Field.SubField }}
    content = re.sub(r'\{\{\s*\.(\w+)\.(\w+)\.(\w+)\s*\}\}', r'{{ \1.\2.\3 }}', content)
    
    # Convert {{if .Variable}} to {% if Variable %}
    content = re.sub(r'\{\{\s*if\s+\.(\w+)\s*\}\}', r'{% if \1 %}', content)
    
    # Convert {{if .Object.Field}} to {% if Object.Field %}
    content = re.sub(r'\{\{\s*if\s+\.(\w+)\.(\w+)\s*\}\}', r'{% if \1.\2 %}', content)
    
    # Convert {{if eq .Variable "value"}} to {% if Variable == "value" %}
    content = re.sub(r'\{\{\s*if\s+eq\s+\.(\w+)\s+"([^"]+)"\s*\}\}', r'{% if \1 == "\2" %}', content)
    content = re.sub(r'\{\{\s*if\s+eq\s+\.(\w+)\.(\w+)\s+"([^"]+)"\s*\}\}', r'{% if \1.\2 == "\3" %}', content)
    
    # Convert {{if ne .Variable "value"}} to {% if Variable != "value" %}
    content = re.sub(r'\{\{\s*if\s+ne\s+\.(\w+)\s+"([^"]+)"\s*\}\}', r'{% if \1 != "\2" %}', content)
    
    # Convert {{if gt .Variable value}} to {% if Variable > value %}
    content = re.sub(r'\{\{\s*if\s+gt\s+\.(\w+)\s+(\d+)\s*\}\}', r'{% if \1 > \2 %}', content)
    
    # Convert {{if lt .Variable value}} to {% if Variable < value %}
    content = re.sub(r'\{\{\s*if\s+lt\s+\.(\w+)\s+(\d+)\s*\}\}', r'{% if \1 < \2 %}', content)
    
    # Convert {{else}} to {% else %}
    content = re.sub(r'\{\{\s*else\s*\}\}', r'{% else %}', content)
    
    # Convert {{else if .Variable}} to {% elif Variable %}
    content = re.sub(r'\{\{\s*else\s+if\s+\.(\w+)\s*\}\}', r'{% elif \1 %}', content)
    
    # Convert {{range .Items}} to {% for item in Items %}
    content = re.sub(r'\{\{\s*range\s+\.(\w+)\s*\}\}', r'{% for item in \1 %}', content)
    
    # Convert {{range $key, $value := .Items}} to {% for key, value in Items.items %}
    content = re.sub(r'\{\{\s*range\s+\$(\w+),\s*\$(\w+)\s*:=\s*\.(\w+)\s*\}\}', r'{% for \1, \2 in \3.items %}', content)
    
    # Convert {{range $index, $item := .Items}} to {% for item in Items %}
    content = re.sub(r'\{\{\s*range\s+\$\w+,\s*\$(\w+)\s*:=\s*\.(\w+)\s*\}\}', r'{% for \1 in \2 %}', content)
    
    # Convert $.Variable to Variable (parent context)
    content = re.sub(r'\{\{\s*\$\.(\w+)\s*\}\}', r'{{ \1 }}', content)
    
    # Convert {{end}} after if to {% endif %}
    # Convert {{end}} after range to {% endfor %}
    # This is tricky, we need context. For now, convert all {{end}} to {% endif %}
    # and manually fix range endings if needed
    content = re.sub(r'\{\{\s*end\s*\}\}', r'{% endif %}', content)
    
    # Convert {{template "name" .}} to {% include "name" %}
    content = re.sub(r'\{\{\s*template\s+"([^"]+)"\s*\.\s*\}\}', r'{% include "\1" %}', content)
    
    # Convert {{block "name" .}}...{{end}} to {% block name %}...{% endblock %}
    content = re.sub(r'\{\{\s*block\s+"([^"]+)"\s*\.\s*\}\}', r'{% block \1 %}', content)
    # Note: {{end}} for blocks should be {% endblock %} but we already converted to {% endif %}
    # This needs manual fixing in some cases
    
    # Convert {{- and -}} (whitespace trim) - Pongo2 uses same syntax
    content = re.sub(r'\{\{-', r'{%-', content)
    content = re.sub(r'-\}\}', r'-%}', content)
    
    # Convert | funcName to |filtername
    # e.g., {{.Price | formatPrice}} to {{ Price|formatPrice }}
    content = re.sub(r'\{\{\s*\.(\w+)\s*\|\s*(\w+)\s*\}\}', r'{{ \1|\2 }}', content)
    
    # Clean up any double spaces in braces
    content = re.sub(r'\{\{\s+', r'{{ ', content)
    content = re.sub(r'\s+\}\}', r' }}', content)
    
    return content

def process_file(filepath):
    """Process a single file"""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Check if file has Go template syntax
    if '{{.' in content or '{{if' in content or '{{range' in content or '{{template' in content or '{{block' in content:
        original = content
        converted = convert_go_to_pongo2(content)
        
        if original != converted:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(converted)
            return True
    return False

def main():
    template_dir = 'web/templates'
    converted_count = 0
    
    for root, dirs, files in os.walk(template_dir):
        for file in files:
            if file.endswith('.html'):
                filepath = os.path.join(root, file)
                if process_file(filepath):
                    print(f"✅ Converted: {filepath}")
                    converted_count += 1
    
    print(f"\n📊 Converted {converted_count} files")

if __name__ == '__main__':
    main()
