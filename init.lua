---
---gurugeek 2 July 2018
---




box.cfg {  listen              = '3301'}
if box.space.todo == nil then
    box.schema.create_space('todo')
    box.space.todo:format({{name='id', type='unsigned'}, {name='title', type='string'}, {name='completed', type='boolean'}, {name='created', type='integer'},{name='owner', type='string'}})
    box.space.todo:create_index('primary', {type='hash', unique=true, parts={1, 'unsigned'}})
    box.space.todo:create_index('created', {type='tree', parts={4, 'integer'}})
    box.space.todo:create_index('owner', {type='tree', parts={5, 'string'}, unique=false})
end

box.schema.user.create('todo', {password='test', if_not_exists=true})
box.schema.user.grant('todo', 'read,write', 'space', 'todo', {if_not_exists=true})
