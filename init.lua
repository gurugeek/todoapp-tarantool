---
--- Created by a.mayorskiy.
--- DateTime: 23.02.18 20:00
---

box.schema.user.create('todo', {password='test', if_not_exists=true})
box.schema.user.grant('todo', 'read,write', 'space', 'todo', {if_not_exists=true})

if box.space.todo == nil then
    box.schema.create_space('todo')
    box.space.todo:format({{name='id', type='unsigned'}, {name='title', type='string'}, {name='completed', type='boolean'}, {name='created', type='integer'},{name='owner', type='string'}})
    box.space.todo:create_index('primary', {type='hash', unique=true, parts={1, 'unsigned'}})
    box.space.todo:create_index('created', {type='tree', parts={4, 'integer'}})
    box.space.todo:create_index('owner', {type='tree', parts={5, 'string'}, unique=false})
end