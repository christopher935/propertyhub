#!/usr/bin/env python3
import os
import re
import sys

def convert_go_to_pongo2(content, base_template="layouts/admin-base.html"):
    lines = content.split('\n')
    blocks = {}
    current_block = None
    current_block_content = []
    
    i = 0
    while i < len(lines):
        line = lines[i]
        
        if line.strip().startswith('{{template "admin-base"') or line.strip().startswith('{{template "consumer-base"'):
            if 'consumer-base' in line:
                base_template = "layouts/consumer-base.html"
            i += 1
            continue
        
        define_match = re.match(r'\{\{define\s+"([^"]+)"\s*\}\}', line.strip())
        if define_match:
            if current_block:
                blocks[current_block] = '\n'.join(current_block_content)
            current_block = define_match.group(1)
            current_block_content = []
            i += 1
            continue
        
        if line.strip() == '{{end}}':
            if current_block:
                blocks[current_block] = '\n'.join(current_block_content)
                current_block = None
                current_block_content = []
            i += 1
            continue
        
        if current_block:
            current_block_content.append(line)
        
        i += 1
    
    block_mapping = {
        'additional-css': 'additional_css',
        'head': 'head',
        'layout-data': 'layout_data',
        'content': 'content',
        'scripts': 'scripts',
        'title': 'title',
    }
    
    result = [f'{{% extends "{base_template}" %}}']
    result.append('')
    
    for old_name, new_name in block_mapping.items():
        if old_name in blocks:
            block_content = blocks[old_name]
            block_content = convert_template_syntax(block_content)
            result.append(f'{{% block {new_name} %}}')
            result.append(block_content)
            result.append('{% endblock %}')
            result.append('')
    
    return '\n'.join(result)


def convert_template_syntax(content):
    content = re.sub(r'\{\{\.([A-Za-z_][A-Za-z0-9_\.]*)\}\}', r'{{ \1 }}', content)
    content = re.sub(r'\{\{if\s+\.([^}]+)\}\}', r'{% if \1 %}', content)
    content = re.sub(r'\{\{if\s+eq\s+\.([^\s]+)\s+"([^"]+)"\}\}', r'{% if \1 == "\2" %}', content)
    content = re.sub(r'\{\{if\s+ne\s+\.([^\s]+)\s+"([^"]+)"\}\}', r'{% if \1 != "\2" %}', content)
    content = re.sub(r'\{\{else\}\}', r'{% else %}', content)
    content = re.sub(r'\{\{else if\s+\.([^}]+)\}\}', r'{% elif \1 %}', content)
    content = re.sub(r'\{\{end\}\}', r'{% endif %}', content)
    content = re.sub(r'\{\{range\s+\.([A-Za-z_][A-Za-z0-9_\.]*)\}\}', r'{% for item in \1 %}', content)
    content = re.sub(r'\{\{range\s+\$([a-z]+),\s*\$([a-z]+)\s*:=\s*\.([A-Za-z_][A-Za-z0-9_\.]*)\}\}', r'{% for \2 in \3 %}', content)
    content = content.replace('{{end}}', '{% endfor %}')
    content = re.sub(r'\{\{template\s+"shared/components/([^"]+)"\s+\.\}\}', r'{% include "partials/\1" %}', content)
    content = re.sub(r'\{\{template\s+"shared/includes/([^"]+)"\s+\.\}\}', r'{% include "partials/\1" %}', content)
    content = re.sub(r'\{\{template\s+"([^"]+)"\s+\.\}\}', r'{% include "\1" %}', content)
    content = re.sub(r'\{\{\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*formatPrice\}\}', r'{{ \1|formatPrice }}', content)
    content = re.sub(r'\{\{\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*formatDate\}\}', r'{{ \1|formatDate }}', content)
    content = re.sub(r'\{\{\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*formatNumber\}\}', r'{{ \1|formatNumber }}', content)
    content = re.sub(r'\{\{\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*safeHTML\}\}', r'{{ \1|safe }}', content)
    
    return content


def process_file(filepath, base_template="layouts/admin-base.html"):
    with open(filepath, 'r') as f:
        content = f.read()
    
    if '{% extends' in content:
        print(f"  Skipping {filepath} - already converted")
        return False
    
    if '{{template "admin-base"' not in content and '{{template "consumer-base"' not in content:
        if '{{define "' in content:
            print(f"  Converting standalone template: {filepath}")
            new_content = convert_standalone_template(content, base_template)
        else:
            print(f"  Skipping {filepath} - not a page template")
            return False
    else:
        new_content = convert_go_to_pongo2(content, base_template)
    
    with open(filepath, 'w') as f:
        f.write(new_content)
    
    print(f"  Converted: {filepath}")
    return True


def convert_standalone_template(content, base_template):
    content = convert_template_syntax(content)
    define_pattern = r'\{\{define\s+"[^"]+"\s*\}\}'
    content = re.sub(define_pattern, '', content)
    content = content.replace('{{end}}', '')
    return content.strip()


def main():
    base_dir = "/project/workspace/christopher935/propertyhub/web/templates"
    
    admin_pages = os.path.join(base_dir, "admin/pages")
    for filename in os.listdir(admin_pages):
        if filename.endswith('.html') and not filename.endswith('.backup'):
            filepath = os.path.join(admin_pages, filename)
            process_file(filepath, "layouts/admin-base.html")
    
    consumer_pages = os.path.join(base_dir, "consumer/pages")
    for filename in os.listdir(consumer_pages):
        if filename.endswith('.html'):
            filepath = os.path.join(consumer_pages, filename)
            process_file(filepath, "layouts/consumer-base.html")
    
    auth_pages = os.path.join(base_dir, "auth/pages")
    for filename in os.listdir(auth_pages):
        if filename.endswith('.html'):
            filepath = os.path.join(auth_pages, filename)
            content = open(filepath).read()
            if '{{template' in content or '{{define' in content:
                new_content = convert_standalone_template(content, None)
                with open(filepath, 'w') as f:
                    f.write(new_content)
                print(f"  Converted standalone auth template: {filepath}")
    
    error_pages = os.path.join(base_dir, "errors/pages")
    for filename in os.listdir(error_pages):
        if filename.endswith('.html'):
            filepath = os.path.join(error_pages, filename)
            content = open(filepath).read()
            if '{{template' in content or '{{define' in content:
                new_content = convert_standalone_template(content, None)
                with open(filepath, 'w') as f:
                    f.write(new_content)
                print(f"  Converted standalone error template: {filepath}")
    
    other_dirs = ["commissions/pages", "leads/pages"]
    for dir_path in other_dirs:
        full_path = os.path.join(base_dir, dir_path)
        if os.path.exists(full_path):
            for filename in os.listdir(full_path):
                if filename.endswith('.html'):
                    filepath = os.path.join(full_path, filename)
                    process_file(filepath, "layouts/admin-base.html")

    print("\nConversion complete!")


if __name__ == "__main__":
    main()
